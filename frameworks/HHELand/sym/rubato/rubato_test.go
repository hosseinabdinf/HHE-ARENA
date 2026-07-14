package rubato

import (
	"fmt"
	"os"
	"testing"

	"HHELand"
	"HHELand/utils"
)

func testString(opName string, p Parameter) string {
	return fmt.Sprintf("%s/BlockSize=%d/Modulus=%d/Rounds=%d/Sigma=%f",
		opName, p.GetBlockSize(), p.GetModulus(), p.GetRounds(), p.GetSigma())
}

func TestRubato(t *testing.T) {
	logger := utils.NewLogger(utils.DEBUG)
	for _, tc := range TestsVector {
		fmt.Println(testString("Rubato", tc.Params))
		rubatoCipher := NewRubato(tc.Key, tc.Params)
		encryptor := rubatoCipher.NewEncryptor()
		var ciphertext HHELand.Ciphertext

		t.Run("RubatoEncryptionTest", func(t *testing.T) {
			ciphertext = encryptor.Encrypt(tc.Plaintext)
			logger.PrintMemUsage("RubatoEncryptionTest")
		})

		t.Run("RubatoDecryptionTest", func(t *testing.T) {
			encryptor.Decrypt(ciphertext)
		})

		logger.PrintDataLen(tc.Key)
		logger.PrintDataLen(ciphertext)
	}
}

// TestNbRubato new benchmarking test for rubato
func TestNbRubato(t *testing.T) {
	t.Run("HHE-ARENA:Rubato", func(t *testing.T) {
		utils.PrintHeader("Rubato SKE Benchmark")
		variants := map[string]int{"128S": 0, "128M": 1, "128L": 2}
		variant := os.Getenv("HHE_BENCH_VARIANT")
		if variant == "" {
			variant = "128L"
		}
		index, ok := variants[variant]
		if !ok {
			t.Fatalf("unknown HHE_BENCH_VARIANT %q", variant)
		}
		tCase := TestsVector[index]

		fmt.Println(testString("Rubato", tCase.Params))

		rubatoCipher := NewRubato(tCase.Key, tCase.Params)
		encryptor := rubatoCipher.NewEncryptor()

		utils.BenchmarkIter("SKE.Enc()", 100, func() {
			_ = encryptor.Encrypt(tCase.Plaintext)
		})
	})
}
