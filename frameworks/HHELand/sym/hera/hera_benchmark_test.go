package hera

import (
	"HHELand"
	"HHELand/utils"
	"fmt"
	"sync"
	"testing"
)

func BenchmarkHera(b *testing.B) {
	for _, tc := range TestVector {
		benchmarkHera(&tc, b)
	}
}

func benchmarkHera(tc *TestContext, b *testing.B) {
	fmt.Println(testString("HERA", tc.Params))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}
	var heraCipher Hera
	var encryptor Encryptor
	var ciphertext HHELand.Ciphertext
	runArenaBenchmark(b, "HERA/NewHera", func() {
		heraCipher = NewHera(tc.Key, tc.Params)
	})
	runArenaBenchmark(b, "HERA/NewEncryptor", func() {
		encryptor = heraCipher.NewEncryptor()
	})
	runArenaBenchmark(b, "HERA/Encrypt", func() {
		ciphertext = encryptor.Encrypt(tc.Plaintext)
	})
	runArenaBenchmark(b, "HERA/Decrypt", func() {
		encryptor.Decrypt(ciphertext)
	})
}

// runArenaBenchmark lets go test retain its benchmark discovery interface while
// the shared harness owns warm-up, iteration, timing, memory, and JSONL output.
func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 100, target) })
		b.ReportMetric(0, "ns/op")
	})
}
