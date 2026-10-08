package pasta

import (
	"HHELand/utils"
	"encoding/binary"
	"fmt"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func BenchmarkPasta3(b *testing.B) {
	for _, tc := range selectedPastaCases(b, "pasta3", pasta3TestVector) {
		benchHEPasta(tc, b)
	}
}

func BenchmarkPasta4(b *testing.B) {
	for _, tc := range selectedPastaCases(b, "pasta4", pasta4TestVector) {
		benchHEPasta(tc, b)
	}
}

func selectedPastaCases(b *testing.B, cipher string, cases []TestContext) []TestContext {
	selected := os.Getenv("HHE_BENCH_VARIANT")
	if selected == "" {
		return cases
	}
	prefix := "hheland-" + cipher + "-p"
	if !strings.HasPrefix(selected, prefix) {
		otherCipher := "pasta3"
		if cipher == "pasta3" {
			otherCipher = "pasta4"
		}
		if strings.HasPrefix(selected, "hheland-"+otherCipher+"-") {
			return nil
		}
		b.Fatalf("unsupported HHE_BENCH_VARIANT %q for HHE Pasta", selected)
	}
	modulus, err := strconv.ParseUint(strings.TrimPrefix(selected, prefix), 10, 64)
	if err != nil {
		b.Fatalf("invalid HHE_BENCH_VARIANT %q for HHE Pasta", selected)
	}
	for _, tc := range cases {
		if tc.SymParams.Modulus == modulus {
			return []TestContext{tc}
		}
	}
	b.Fatalf("HHE_BENCH_VARIANT %q has no matching %s parameter case", selected, cipher)
	return nil
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
