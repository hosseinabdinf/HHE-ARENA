/*
  Example program that encrypts a plaintext file using Fasta and writes the
  ciphertext to file. Then it reads in the ciphertext file and decrypts it to
  verify that we get back the same plaintext. Key is generated at random on every
  run.
*/

#include "Fasta.h"
#include "utils.h"
#include <stdio.h>
#include <string.h>

struct plaintext *readPlaintextFile(const char *filename) {
  /* Reads in plaintext from the given file and returns it in a plaintext
   * object. */
  FILE *fp;
  struct plaintext *PT;
  int i, j;

  fp = fopen(filename, "r");
  if (fp == NULL) {
    printf("Could not open the file %s\n", filename);
    exit(0);
  }

  PT = (struct plaintext *)malloc(sizeof(struct plaintext));
  fscanf(fp, "%d\n", &PT->numBlocks);
  PT->block =
      (struct BitVector **)malloc(PT->numBlocks * sizeof(struct BitVector *));
  for (i = 0; i < PT->numBlocks; ++i) {
    PT->block[i] = NewBitVector(1645);
    for (j = 51; j >= 0; --j)
      fscanf(fp, "%8x ", PT->block[i]->values + j);
    fscanf(fp, "\n");
  }
  fclose(fp);

  return PT;
}

struct ciphertext *readCiphertextFile(const char *filename) {
  /* Reads in ciphertext from the given file and returns it in a ciphertext
   * object. */
  FILE *fp;
  struct ciphertext *CT;
  int i, j;

  fp = fopen(filename, "r");
  if (fp == NULL) {
    printf("Could not open the file %s\n", filename);
    exit(0);
  }

  CT = (struct ciphertext *)malloc(sizeof(struct ciphertext));
  fscanf(fp, "%d\n", &CT->numBlocks);
  CT->block =
      (struct BitVector **)malloc(CT->numBlocks * sizeof(struct BitVector *));
  for (i = 0; i < CT->numBlocks; ++i) {
    CT->block[i] = NewBitVector(1645);
    for (j = 51; j >= 0; --j)
      fscanf(fp, "%8x ", CT->block[i]->values + j);
    fscanf(fp, "\n");
  }
  fclose(fp);

  return CT;
}

void printPlaintextFile(struct plaintext *PT, const char *filename) {
  /* Prints the plaintext PT to a file with the given name. */
  FILE *fp;
  int i, j;

  fp = fopen(filename, "w");
  fprintf(fp, "%d\n", PT->numBlocks);
  for (i = 0; i < PT->numBlocks; ++i) {
    for (j = 51; j >= 0; --j)
      fprintf(fp, "%08x ", PT->block[i]->values[j]);
    fprintf(fp, "\n");
  }
  fclose(fp);
}

void printCiphertextFile(struct ciphertext *CT, const char *filename) {
  /* Prints the ciphertext CT to a file with the given name. */
  FILE *fp;
  int i, j;

  fp = fopen(filename, "w");
  fprintf(fp, "%d\n", CT->numBlocks);
  for (i = 0; i < CT->numBlocks; ++i) {
    for (j = 51; j >= 0; --j)
      fprintf(fp, "%08x ", CT->block[i]->values[j]);
    fprintf(fp, "\n");
  }
  fclose(fp);
}


static void deletePlaintext(struct plaintext *value) {
  if (value == NULL) return;
  for (int i = 0; i < value->numBlocks; ++i) DeleteBitVector(value->block[i]);
  free(value->block);
  free(value);
}

static void deleteCiphertext(struct ciphertext *value) {
  if (value == NULL) return;
  for (int i = 0; i < value->numBlocks; ++i) DeleteBitVector(value->block[i]);
  free(value->block);
  free(value);
}

int main(int argc, char *argv[]) {
  (void)argc;
  (void)argv;
  struct plaintext *ptxt, *decrypted = NULL;
  struct ciphertext *ctxt = NULL;
  struct BitVector *K;

  K = RandomBitVector(329);
  ptxt = readPlaintextFile("originalPlaintext.txt");
  printf("> Reading Plaintext (numblock: %d, lenght: %d)\n", ptxt->numBlocks,
         ptxt->block[0]->length);

  const char *requested_operation = getenv("HHE_BENCH_OPERATION");
  int run_encrypt = requested_operation == NULL || requested_operation[0] == '\0' ||
                    strcmp(requested_operation, "ske_encrypt") == 0;
  int run_decrypt = requested_operation == NULL || requested_operation[0] == '\0' ||
                    strcmp(requested_operation, "ske_decrypt") == 0;
  if (!run_encrypt && !run_decrypt) {
    fprintf(stderr, "Unsupported Fasta operation: %s\n", requested_operation);
    deletePlaintext(ptxt);
    DeleteBitVector(K);
    return 2;
  }

  if (run_encrypt) {
    PrintHeader("SYM Encryption");
    HHE_BENCHMARK_ITER_HOOKS(
        "ske_encrypt", 1,
        if (ctxt != NULL) { deleteCiphertext(ctxt); ctxt = NULL; },
        ctxt = encrypt(ptxt, K),
        (void)0);
  }

  if (run_decrypt) {
    if (ctxt == NULL) ctxt = encrypt(ptxt, K);
    PrintHeader("SYM Decryption");
    HHE_BENCHMARK_ITER_HOOKS(
        "ske_decrypt", 1,
        if (decrypted != NULL) { deletePlaintext(decrypted); decrypted = NULL; },
        decrypted = decrypt(ctxt, K),
        (void)0);
  }

  // printCiphertextFile(ctxt, "ciphertext.txt");
  // printf("> Encryption of plaintext written to ciphertext.txt\n");
  // printPlaintextFile(decrypted, "decryptedPlaintext.txt");
  // printf("Ciphertext decrypted and written to decryptedPlaintext.txt\n");

  deletePlaintext(decrypted);
  deleteCiphertext(ctxt);
  deletePlaintext(ptxt);
  DeleteBitVector(K);
  return 0;
}

void printTimeNS() {}
