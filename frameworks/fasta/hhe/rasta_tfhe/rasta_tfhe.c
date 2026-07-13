/* Implementation of Rasta using TFHE.
   Used only for timing purposes, key and constant additions are not
   implemented. */
#include <stdio.h>
#include <stdlib.h>
#include <tfhe/tfhe.h>
#include <tfhe/tfhe_io.h>
#include <time.h>

#include "utils.h"

int BLOCK_SIZE = 351;
TFheGateBootstrappingParameterSet *context;
TFheGateBootstrappingSecretKeySet *secret_key;
int ***Mlist;
double timeChi = 0, timeLinLayer = 0;

void fillMlist(int n_rounds) {
  /* Reads in n_rounds matrices from files named "matrix_a.txt" (where a is
   * 1,2,...,). */
  int i, j, m;
  FILE *fp;
  char filename[20];

  Mlist = (int ***)malloc(n_rounds * sizeof(int **));
  for (m = 0; m < n_rounds; ++m) {
    sprintf(filename, "matrix_%d.txt", m + 1);
    fp = fopen(filename, "r");
    Mlist[m] = (int **)malloc(BLOCK_SIZE * sizeof(int *));
    for (i = 0; i < BLOCK_SIZE; ++i) {
      Mlist[m][i] = (int *)malloc(BLOCK_SIZE * sizeof(int));
      for (j = 0; j < BLOCK_SIZE; ++j)
        fscanf(fp, "%d", Mlist[m][i] + j);
    }
    fclose(fp);
  }
}


void freeMlist(int n_rounds) {
  if (Mlist == NULL) return;
  for (int m = 0; m < n_rounds; ++m) {
    for (int i = 0; i < BLOCK_SIZE; ++i) free(Mlist[m][i]);
    free(Mlist[m]);
  }
  free(Mlist);
  Mlist = NULL;
}

LweSample *Chi(LweSample *input,
               const TFheGateBootstrappingCloudKeySet cloud_key) {
  /* Computes and returns Chi(input). */
  int i;
  LweSample *tmp, *output;

  output = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);
  tmp = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);

  for (i = 0; i < BLOCK_SIZE; ++i) {
    bootsANDYN(tmp + i, input + ((i + 2) % BLOCK_SIZE),
               input + ((i + 1) % BLOCK_SIZE), &cloud_key);
    bootsXOR(output + i, tmp + i, input + i, &cloud_key);
  }

  delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, tmp);
  return output;
}

LweSample *
matrixMultiplicationPlain(LweSample *input, int **M,
                          const TFheGateBootstrappingCloudKeySet cloud_key) {
  /* Computes and returns input*M in the straigh-forward way. */
  int i, j;
  LweSample *output;

  output = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);
  for (i = 0; i < BLOCK_SIZE; ++i)
    bootsSymEncrypt(output + i, 0, secret_key);
  // output initialized
  for (j = 0; j < BLOCK_SIZE; ++j) {
    for (i = 0; i < BLOCK_SIZE; ++i) {
      if (M[i][j] != 0)
        bootsXOR(output + j, output + j, input + i, &cloud_key);
    }
  }

  return output;
}

LweSample *
matrixMultiplicationM4R(LweSample *input, int **M,
                        const TFheGateBootstrappingCloudKeySet cloud_key) {
  /* Computes and returns input*M using the method of the four russians.  For
   * BLOCK_SIZE=351, t=5 is optimal. */
  int i, j, t = 5 /*optimal*/, nTables, tableSize, tableIndex, p, rest;
  LweSample *output, **ciphertextTable;

  // Create lookup-tables
  nTables = BLOCK_SIZE / t; // will be 70 when BLOCKSIZE=351 and t=5
  tableSize = 1 << t;
  ciphertextTable = (LweSample **)malloc(nTables * sizeof(LweSample *));
  for (i = 0; i < nTables; ++i) {
    ciphertextTable[i] =
        new_gate_bootstrapping_ciphertext_array(tableSize, context);
    for (j = 0; j < tableSize; ++j)
      bootsSymEncrypt(ciphertextTable[i] + j, 0, secret_key);
    // ciphertextTable[i] initialized
    for (tableIndex = 1; tableIndex < tableSize;
         ++tableIndex) { // nothing to add when tableIndex is 0
      for (p = 0; p < t; ++p) {
        if ((1 << p) & tableIndex) // must add ciphertext p of input block i
                                   // onto current table index
          bootsXOR(ciphertextTable[i] + tableIndex,
                   ciphertextTable[i] + tableIndex, input + (t * i + p),
                   &cloud_key);
      }
    } // ciphertextTable[i] initialized
  } // all look-up tables initialized

  output = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);
  for (i = 0; i < BLOCK_SIZE; ++i)
    bootsSymEncrypt(output + i, 0, secret_key);
  // output initialized

  // Do the actual multiplication
  rest = BLOCK_SIZE % t;
  for (j = 0; j < BLOCK_SIZE; ++j) { // for every column...
    for (i = 0; i < nTables; ++i) {  //...look up every block of t rows
      tableIndex = 0;
      for (p = 0; p < t; ++p) {
        if (M[(t * i) + p][j] != 0)
          tableIndex ^= (1 << p);
      } // tableIndex indicates the element to add to output
      if (tableIndex != 0)
        bootsXOR(output + j, output + j, ciphertextTable[i] + tableIndex,
                 &cloud_key);
    }
    // adding the last BLOCK_SIZE mod t elements of input
    for (p = BLOCK_SIZE - rest; p < BLOCK_SIZE; ++p) {
      if (M[p][j] != 0)
        bootsXOR(output + j, output + j, input + p, &cloud_key);
    } // column j fully processed
  }

  for (i = 0; i < nTables; ++i)
    delete_gate_bootstrapping_ciphertext_array(tableSize, ciphertextTable[i]);
  free(ciphertextTable);
  return output;
}

LweSample *RastaRound(LweSample *input, int r,
                      const TFheGateBootstrappingCloudKeySet cloud_key) {
  /* One Rasta round, with Chi and matrix multiplication (no key or constant
   * addition). */
  LweSample *afterChi = Chi(input, cloud_key);
  LweSample *output =
      matrixMultiplicationM4R(afterChi, Mlist[r], cloud_key);
  delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, afterChi);
  return output;
}

int main(int argc, char *argv[]) {
  LweSample *roundInput, *roundOutput = NULL;
  int i, c;

  PrintHeader("Initializing TFHE context");
  HHE_BENCHMARK_ITER_HOOKS(
      "he_context_setup", 1,
      {
        if (context != NULL) {
          delete_gate_bootstrapping_parameters(context);
          context = NULL;
        }
      },
      context = new_default_gate_bootstrapping_parameters(128),
      (void)0);

  PrintHeader("Generating secret key");
  secret_key = new_random_gate_bootstrapping_secret_keyset(context);
  const TFheGateBootstrappingCloudKeySet cloud_key = secret_key->cloud;
  printf("-> done!\n");

  fillMlist(7);

  roundInput = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);
  roundOutput = new_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, context);

  printf("-> Plaintext: [");
  for (i = 0; i < BLOCK_SIZE; ++i) {
    c = rand() % 2;
    printf("%d ", c);
    // encrypt plaintexts
    bootsSymEncrypt(roundInput + i, c, secret_key);
  }
  printf("]\n");

  PrintHeader("Starting TFHE Rasta Transciphering");
  HHE_BENCHMARK_ITER_HOOKS(
      "hhe_transcipher", 1,
      {
        if (roundOutput != NULL) {
          delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, roundOutput);
          roundOutput = NULL;
        }
      },
      {
        roundOutput = matrixMultiplicationM4R(roundInput, Mlist[0], cloud_key);
        for (i = 1; i <= 6; ++i) {
          LweSample *next = RastaRound(roundOutput, i, cloud_key);
          delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, roundOutput);
          roundOutput = next;
        }
      },
      (void)0);

  delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, roundOutput);
  delete_gate_bootstrapping_ciphertext_array(BLOCK_SIZE, roundInput);
  delete_gate_bootstrapping_secret_keyset(secret_key);
  delete_gate_bootstrapping_parameters(context);
  freeMlist(7);
  return 0;
}
