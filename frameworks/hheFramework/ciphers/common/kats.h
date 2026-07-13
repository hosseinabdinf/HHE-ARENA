#pragma once

#include "benchmark_kat.h"

#include <algorithm>
#include <cstddef>
#include <iterator>
#include <vector>

#include "Cipher.h"
#include "matrix.h"
#include "utils.h"

template <class T>
class KnownAnswerTest {
 public:
  enum Testcase { MAT, DEC, USE_CASE, PREP };

 private:
  static constexpr bool PRESET_SEED = true;
  utils::SeededRandom rand;
  std::vector<uint8_t> key;
  std::vector<uint8_t> plaintext;
  std::vector<uint8_t> ciphertext_expected;
  Testcase tc;
  size_t N;
  size_t BITSIZE;

 public:
  KnownAnswerTest(std::vector<uint8_t> key, std::vector<uint8_t> plaintext,
                  std::vector<uint8_t> ciphertext_expected, Testcase tc = DEC,
                  size_t N = 4, size_t bitsize = 16)
      : rand(PRESET_SEED),
        key(key),
        plaintext(plaintext),
        ciphertext_expected(ciphertext_expected),
        tc(tc),
        N(N),
        BITSIZE(bitsize) {}

  KnownAnswerTest(const uint8_t *key, size_t key_size, const uint8_t *plaintext,
                  size_t plaintext_size, const uint8_t *ciphertext_expected,
                  size_t ciphertext_size, Testcase tc = DEC, size_t N = 4,
                  size_t bitsize = 16)
      : rand(PRESET_SEED),
        key(key, key + key_size),
        plaintext(plaintext, plaintext + plaintext_size),
        ciphertext_expected(ciphertext_expected,
                            ciphertext_expected + ciphertext_size),
        tc(tc),
        N(N),
        BITSIZE(bitsize) {}

  bool mat_test() {
    const uint64_t MASK = (1ULL << BITSIZE) - 1;
    std::cout << "> N = " << N << std::endl;
    std::cout << "> BitSize = " << BITSIZE << std::endl;
    std::cout << "-=== Getting random elements ===-" << std::endl;

    // random matrix
    matrix::matrix m(N);
    for (size_t i = 0; i < N; i++) {
      m[i].reserve(N);
      for (size_t j = 0; j < N; j++) {
        m[i].push_back(rand.random_uint64() & MASK);
      }
    }

    // random bias
    matrix::vector b;
    b.reserve(N);
    for (size_t i = 0; i < N; i++) {
      b.push_back(rand.random_uint64() & MASK);
    }

    // random input vector
    matrix::vector vi;
    vi.reserve(N);
    for (size_t i = 0; i < N; i++) {
      vi.push_back(rand.random_uint64() & MASK);
    }
    std::cout << "-> done! \n";

    std::cout << "-=== Computing in plain ===-" << std::endl;
    matrix::vector vo;
    matrix::affine(vo, m, vi, b, BITSIZE);
    std::cout << "-> done! \n";
    return true;
  }

  bool dec_test() {

    T cipher(key);
    std::cout << "-=== Final decrypt ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("ske_decrypt");

    std::vector<uint8_t> plain =
        cipher.decrypt(ciphertext_expected, cipher.get_cipher_size_bits());

    HHE_KAT_OPERATION_END();
    std::cout << "-> done! \n";

    if (plain != plaintext) {
      std::cerr << cipher.get_cipher_name() << " KATS failed!\n";
      utils::print_vector("key:      ", key, std::cerr);
      utils::print_vector("ciphertext", ciphertext_expected, std::cerr);
      utils::print_vector("plain:    ", plaintext, std::cerr);
      utils::print_vector("got:      ", plain, std::cerr);
      return false;
    }
    return true;
  }

  bool test() {

    const uint64_t MASK = (1ULL << BITSIZE) - 1;
    std::cout << "> N = " << N << std::endl;
    std::cout << "> BitSize = " << BITSIZE << std::endl;
    std::cout << "-=== Getting random elements ===-" << std::endl;

    // random matrix
    matrix::matrix m(N);
    for (size_t i = 0; i < N; i++) {
      m[i].reserve(N);
      for (size_t j = 0; j < N; j++) {
        m[i].push_back(rand.random_uint64() & MASK);
      }
    }

    // random bias
    matrix::vector b;
    b.reserve(N);
    for (size_t i = 0; i < N; i++) {
      b.push_back(rand.random_uint64() & MASK);
    }

    // random input vector
    matrix::vector vi;
    vi.reserve(N);
    for (size_t i = 0; i < N; i++) {
      vi.push_back(rand.random_uint64() & MASK);
    }
    std::cout << "-> done! \n";

    std::cout << "-=== Computing in plain ===-" << std::endl;
    matrix::vector vo, vo_p(N);
    HHE_KAT_OPERATION_BEGIN("plain_evaluate");
    matrix::affine(vo, m, vi, b, BITSIZE);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done! \n";

    std::cout << "-=== Encrypting input ===-" << std::endl;
    T cipher(key);
    std::vector<uint8_t> vi_encoded;
    utils::encode(vi_encoded, vi, BITSIZE);
    HHE_KAT_OPERATION_BEGIN("ske_encrypt");
    std::vector<uint8_t> ciph = cipher.encrypt(vi_encoded, BITSIZE * N);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done! \n";

    std::cout << "-=== Decrypting ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("ske_decrypt");
    std::vector<uint8_t> plain_encoded = cipher.decrypt(ciph, BITSIZE * N);
    HHE_KAT_OPERATION_END();
    matrix::vector plain;
    utils::decode(plain, plain_encoded, BITSIZE);
    std::cout << "-> done! \n";

    std::cout << "\t -=== Computing ===-" << std::endl;

    HHE_KAT_OPERATION_BEGIN("plain_evaluate");
    matrix::affine(vo_p, m, plain, b, BITSIZE);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done! \n";

    if (vo != vo_p) {
      std::cerr << cipher.get_cipher_name() << " KATS failed!\n";
      utils::print_vector("plain:  ", vo, std::cerr);
      utils::print_vector("cipher: ", vo_p, std::cerr);
      return false;
    }
    return true;
  }

  bool prep_test() {

    T cipher(key);
    std::cout << "-=== Preprocessing constants for 1 block ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("preprocess_block");
    cipher.prep_one_block();
    HHE_KAT_OPERATION_END();
    std::cout << "-> done! \n";

    return true;
  }

  virtual bool run_test() {
    switch (tc) {
      case MAT:
        return mat_test();
      case DEC:
        return dec_test();
      case USE_CASE:
        return test();
      case PREP:
        return prep_test();
      default: {
        std::cout << "Testcase not found... " << std::endl;
        return false;
      }
    }
  }
};
