package pasta

import (
	"HHELand/utils"
	"encoding/binary"
	"fmt"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"sync"
	"testing"
)

func BenchmarkPasta3(b *testing.B) {
	for _, tc := range pasta3TestVector {
		benchHEPasta(tc, b)
	}
}

func BenchmarkPasta4(b *testing.B) {
	for _, tc := range pasta4TestVector {
		benchHEPasta(tc, b)
	}
}

func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 1, target) })
		b.ReportMetric(0, "ns/op")
	})
}

func benchHEPasta(tc TestContext, b *testing.B) {
	fmt.Println(testString("PASTA", tc.SymParams))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}

	logger := utils.NewLogger(utils.DEBUG)
	logger.PrintDataLen(tc.Key)

	hePasta := NewHEPasta()
	hePasta.InitParams(tc.Params, tc.SymParams)

	runArenaBenchmark(b, "PASTA/HEKeyGen", func() {
		hePasta.HEKeyGen()
	})

	_ = hePasta.InitFvPasta()

	runArenaBenchmark(b, "PASTA/GaloisKeysGen", func() {
		hePasta.CreateGaloisKeys(len(tc.ExpCipherText))
	})

	runArenaBenchmark(b, "PASTA/EncryptSymKey", func() {
		hePasta.EncryptSymKey(tc.Key)
	})

	nonce := make([]byte, 8)
	binary.BigEndian.PutUint64(nonce, uint64(123456789))

	var fvCiphers []*rlwe.Ciphertext
	runArenaBenchmark(b, "PASTA/Trancipher", func() {
		fvCiphers = hePasta.Trancipher(nonce, tc.ExpCipherText)
	})

	runArenaBenchmark(b, "PASTA/Decrypt", func() {
		hePasta.Decrypt(fvCiphers[0])
	})
}
