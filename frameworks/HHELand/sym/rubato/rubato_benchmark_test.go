package rubato

import (
	"HHELand"
	"HHELand/utils"
	"fmt"
	"sync"
	"testing"
)

func BenchmarkRubato(b *testing.B) {
	for _, tc := range TestsVector {
		benchmarkRubato(&tc, b)
	}
}

func benchmarkRubato(tc *TestContext, b *testing.B) {
	fmt.Println(testString("Rubato", tc.Params))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}
	var cipher Rubato
	var encryptor Encryptor
	var ciphertext HHELand.Ciphertext
	runArenaBenchmark(b, "Rubato/NewRubato", func() { cipher = NewRubato(tc.Key, tc.Params) })
	runArenaBenchmark(b, "Rubato/NewEncryptor", func() { encryptor = cipher.NewEncryptor() })
	runArenaBenchmark(b, "Rubato/Encrypt", func() { ciphertext = encryptor.Encrypt(tc.Plaintext) })
	runArenaBenchmark(b, "Rubato/Decrypt", func() { encryptor.Decrypt(ciphertext) })
}

func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 100, target) })
		b.ReportMetric(0, "ns/op")
	})
}
