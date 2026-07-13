#pragma once

#include "../common/benchmark_kat.h"

#include <algorithm>
#include <cstddef>
#include <iterator>
#include <vector>

#include "../common/utils.h"
#include "Cipher.h"
#include "matrix.h"

template <class T>
class KnownAnswerTestZp {
 public:
  enum Testcase { MAT, DEC, USE_CASE, PREP };

 private:
  static constexpr bool PRESET_SEED = true;
  static constexpr size_t NUM_MATMULS_SQUARES = 3;
  static constexpr bool LAST_SQUARE = false;
  utils::SeededRandom rand;
  std::vector<uint64_t> key;
  std::vector<uint64_t> plaintext;
  std::vector<uint64_t> ciphertext_expected;
  size_t modulus;
  Testcase tc;
  size_t N;

 public:
  KnownAnswerTestZp(std::vector<uint64_t> key, std::vector<uint64_t> plaintext,
                    std::vector<uint64_t> ciphertext_expected, size_t modulus,
                    Testcase tc = DEC, size_t N = 4)
      : rand(PRESET_SEED),
        key(key),
        plaintext(plaintext),
        ciphertext_expected(ciphertext_expected),
        modulus(modulus),
        tc(tc),
        N(N) {}

  KnownAnswerTestZp(const uint64_t *key, size_t key_size,
                    const uint64_t *plaintext, size_t plaintext_size,
                    const uint64_t *ciphertext_expected, size_t ciphertext_size,
                    size_t modulus, Testcase tc = DEC, size_t N = 4)
      : rand(PRESET_SEED),
        key(key, key + key_size),
        plaintext(plaintext, plaintext + plaintext_size),
        ciphertext_expected(ciphertext_expected,
                            ciphertext_expected + ciphertext_size),
        modulus(modulus),
        tc(tc),
        N(N) {}

  bool mat_test() {
    std::cout << "> Num matrices = " << NUM_MATMULS_SQUARES << std::endl;
    std::cout << "> N = " << N << std::endl;
    std::cout << "-=== Getting random elements ===-" << std::endl;

    // random matrices
    std::vector<matrix::matrix> m(NUM_MATMULS_SQUARES);
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      m[r].resize(N);
      for (size_t i = 0; i < N; i++) {
        m[r][i].reserve(N);
        for (size_t j = 0; j < N; j++) {
          m[r][i].push_back(rand.random_uint64() %
                            modulus);  // not cryptosecure ;)
        }
      }
    }

    // random biases
    std::vector<matrix::vector> b(NUM_MATMULS_SQUARES);
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      b[r].reserve(N);
      for (size_t i = 0; i < N; i++) {
        b[r].push_back(rand.random_uint64() % modulus);  // not cryptosecure ;)
      }
    }

    // random input vector
    matrix::vector vi;
    vi.reserve(N);
    for (size_t i = 0; i < N; i++) {
      vi.push_back(rand.random_uint64() % modulus);  // not cryptosecure ;)
    }
    std::cout << "-> done!" << std::endl;
    std::cout << "-=== Computing in plain ===-" << std::endl;

    matrix::vector vo;
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      matrix::affine(vo, m[r], vi, b[r], modulus);
      if (!LAST_SQUARE && r != NUM_MATMULS_SQUARES - 1)
        matrix::square(vi, vo, modulus);
    }
    std::cout << "-> done!" << std::endl;
    return true;
  }

  bool dec_test() {

    T cipher(key, modulus);
    std::cout << "-=== Final decrypt ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("ske_decrypt");
    std::vector<uint64_t> plain = cipher.decrypt(ciphertext_expected);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

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

    std::cout << "> Num matrices = " << NUM_MATMULS_SQUARES << std::endl;
    std::cout << "> N = " << N << std::endl;
    std::cout << "-=== Getting random elements ===-" << std::endl;

    // random matrices
    std::vector<matrix::matrix> m(NUM_MATMULS_SQUARES);
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      m[r].resize(N);
      for (size_t i = 0; i < N; i++) {
        m[r][i].reserve(N);
        for (size_t j = 0; j < N; j++) {
          m[r][i].push_back(rand.random_uint64() %
                            modulus);  // not cryptosecure ;)
        }
      }
    }

    // random biases
    std::vector<matrix::vector> b(NUM_MATMULS_SQUARES);
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      b[r].reserve(N);
      for (size_t i = 0; i < N; i++) {
        b[r].push_back(rand.random_uint64() % modulus);  // not cryptosecure ;)
      }
    }

    // random input vector
    matrix::vector vi;
    vi.reserve(N);
    for (size_t i = 0; i < N; i++) {
      vi.push_back(rand.random_uint64() % modulus);  // not cryptosecure ;)
    }

    std::cout << "-> done!" << std::endl;
    std::cout << "-=== Computing in plain ===-" << std::endl;

    matrix::vector vo, vo_p(N), vi_tmp;
    HHE_KAT_OPERATION_BEGIN("plain_evaluate");
    vi_tmp = vi;
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      matrix::affine(vo, m[r], vi_tmp, b[r], modulus);
      if (!LAST_SQUARE && r != NUM_MATMULS_SQUARES - 1)
        matrix::square(vi_tmp, vo, modulus);
    }
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

    std::cout << "-=== Encrypting input ===-" << std::endl;
    T cipher(key, modulus);
    HHE_KAT_OPERATION_BEGIN("ske_encrypt");
    std::vector<uint64_t> ciph = cipher.encrypt(vi);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

    std::cout << "-=== Decrypting ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("ske_decrypt");
    std::vector<uint64_t> plain = cipher.decrypt(ciph);
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

    std::cout << "-=== Computing ===-" << std::endl;

    HHE_KAT_OPERATION_BEGIN("plain_evaluate");
    for (size_t r = 0; r < NUM_MATMULS_SQUARES; r++) {
      matrix::affine(vo_p, m[r], plain, b[r], modulus);
      if (!LAST_SQUARE && r != NUM_MATMULS_SQUARES - 1)
        matrix::square(plain, vo_p, modulus);
    }
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

    if (vo != vo_p) {
      std::cerr << cipher.get_cipher_name() << " KATS failed!\n";
      utils::print_vector("plain:  ", vo, std::cerr);
      utils::print_vector("cipher: ", vo_p, std::cerr);
      return false;
    }
    return true;
  }

  bool prep_test() {

    T cipher(key, modulus);
    std::cout << "-=== Preprocessing constants for 1 block ===-" << std::endl;
    HHE_KAT_OPERATION_BEGIN("preprocess_block");
    cipher.prep_one_block();
    HHE_KAT_OPERATION_END();
    std::cout << "-> done!" << std::endl;

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
