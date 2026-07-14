package pasta

import (
	"HHELand"
	"HHELand/utils"
	"fmt"
	"sync"
	"testing"
)

func BenchmarkPasta3(b *testing.B) {
	for _, tc := range pasta3TestVector {
		benchmarkPasta(&tc, b)
	}
}

func BenchmarkPasta4(b *testing.B) {
	for _, tc := range pasta4TestVector {
		benchmarkPasta(&tc, b)
	}
}

func benchmarkPasta(tc *TestContext, b *testing.B) {
	fmt.Println(testString("Pasta", tc.Params))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}
	var cipher Pasta
	var encryptor Encryptor
	var ciphertext HHELand.Ciphertext
	runArenaBenchmark(b, "Pasta/NewPasta", func() { cipher = NewPasta(tc.Key, tc.Params) })
	runArenaBenchmark(b, "Pasta/NewEncryptor", func() { encryptor = cipher.NewEncryptor() })
	runArenaBenchmark(b, "Pasta/Encrypt", func() { ciphertext = encryptor.Encrypt(tc.Plaintext) })
	runArenaBenchmark(b, "Pasta/Decrypt", func() { encryptor.Decrypt(ciphertext) })
}

func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 100, target) })
		b.ReportMetric(0, "ns/op")
	})
}
