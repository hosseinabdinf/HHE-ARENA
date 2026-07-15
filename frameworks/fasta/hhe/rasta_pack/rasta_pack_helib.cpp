/*
  Code implementing Rasta with 329-bit block in HElib, using one ciphertext for
  whole block. Used only for timing purposes, addition of constants and keys not
  implemented.
*/

#include <NTL/BasicThreadPool.h>
#include <helib/helib.h>
#include <helib/matmul.h>
#include <stdlib.h>

#include <iostream>
#include <memory>

using namespace std;
using namespace NTL;
using namespace helib;


#include "matrixFromFile.h"
#include "utils.h"

// double timeChi = 0, timeLinLayer = 0;
vector<MatMulFullExec> MAT;

Ctxt Chi(Ctxt input) {
  /* Returns Chi(input). */
  Ctxt i0(input), i1(input), i2(input);
  const EncryptedArray &ea = input.getContext().getEA();

  ea.rotate(i1, 1);
  ea.rotate(i2, 2);

  i1 *= i2;
  i1 += i2;
  i1 += i0;

  return i1;
}

Ctxt LinLayer(Ctxt input, int round) {
  /* Linear layer of Rasta, multiply with random matrix i. */
  Ctxt out(input);

  MAT[round].mul(out);

  return out;
}

Ctxt Round(Ctxt input, int r) {
  /* Returns the state after one round of Rasta. */
  Ctxt out(input);
  int i;
  out = Chi(input);
  out = LinLayer(out, r);
  return out;
}

int main(int argc, char *argv[]) {
  /*  Example of BGV scheme  */
  int i, j;
  // double secs_taken;

  // Plaintext prime modulus
  unsigned long p = 2;
  // Cyclotomic polynomial - defines phi(m)
  unsigned long m = 30269;
  // Hensel lifting (default = 1)
  unsigned long r = 1;
  // Number of bits of the modulus chain
  unsigned long bits = 500;
  // Number of columns of Key-Switching matrix (default = 2 or 3)
  unsigned long c = 2;

  PrintHeader("Initialising Context Object");
  HHE_BENCHMARK_ITER_HOOKS(
      "he_context_setup", 1,
      (void)0,
      { Context context = ContextBuilder<BGV>().m(m).p(p).r(r).bits(bits).c(c).build(); },
      (void)0);
  Context context = ContextBuilder<BGV>().m(m).p(p).r(r).bits(bits).c(c).build();

  // Print the context
  cout << "----------------------------------- \n";
  PrintHeader("HElib Context");
  context.printout();
  cout << "BitQ: " << context.bitSizeOfQ() << endl;
  cout << "P: " << context.getP() << endl;
  cout << "----------------------------------- \n";

  // Secret key management and key-switching setup.
  PrintHeader("Generating key-switching matrices");
  unique_ptr<SecKey> secret_key_ptr;
  HHE_BENCHMARK_ITER_HOOKS(
      "he_key_switch_setup", 1,
      { secret_key_ptr = make_unique<SecKey>(context); secret_key_ptr->GenSecKey(); },
      addSome1DMatrices(*secret_key_ptr),
      (void)0);
  SecKey &secret_key = *secret_key_ptr;

  // Public key management
  // Set the secret key (upcast: SecKey is a subclass of PubKey)
  const PubKey &public_key = secret_key;

  // Get the EncryptedArray of the context
  const EncryptedArray &ea = context.getEA();

  // Get the number of slots (phi(m))
  long nslots = ea.size();
  cout << "-> Number of slots: " << nslots << endl;

  // Reading in random matrices
  char filename[80];
  for (i = 0; i < 7; ++i) {
    sprintf(filename, "matrix_%d.txt", i + 1);
    matrixFromFile nyM(ea, filename);
    MatMulFullExec MMFE(nyM);
    MAT.push_back(MMFE);
  }

  // Create plaintext vector of long with nslots elements each
  Ptxt<BGV> ptxt(context);
  cout << "-> Ptx size:" << ptxt.size() << endl;
  // Fill them with random bits
  for (i = 0; i < ptxt.size(); ++i) ptxt[i] = (long)random() % 2;

  // cout<<"ptxt = "<<ptxt<<endl<<endl;
  // Create a ciphertext object
  Ctxt ctxt(public_key);
  // Encrypt the plaintext using the public_key
  public_key.Encrypt(ctxt, ptxt);

  cout << "-> Bit capacit: " << ctxt.bitCapacity() << endl;

  /* Homomorphic evaluation of Rasta */
  const Ctxt initialCtxt(ctxt);
  PrintHeader("Starting Rasta-Pack Transciphering");
  HHE_BENCHMARK_ITER_HOOKS(
      "hhe_transcipher", 1,
      ctxt = initialCtxt,
      {
        ctxt = LinLayer(ctxt, 0);
        for (i = 1; i <= 6; ++i) ctxt = Round(ctxt, i);
      },
      (void)0);
  cout << "-> Bit capacity: " << ctxt.bitCapacity() << endl;
  return 0;
}
