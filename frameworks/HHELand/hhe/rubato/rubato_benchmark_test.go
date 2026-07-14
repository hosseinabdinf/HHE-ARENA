package rubato

import (
	RtF "HHELand/rtf_integration"
	"HHELand/sym/rubato"
	"HHELand/utils"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
)

func BenchmarkRubato(b *testing.B) {
	for _, tc := range rubato.TestsVector {
		benchHERubato(tc, b)
	}
}

func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 1, target) })
		b.ReportMetric(0, "ns/op")
	})
}

func benchHERubato(tc rubato.TestContext, b *testing.B) {
	fmt.Println(testString("Rubato", tc.Params))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}

	heRubato := NewHERubato()
	heRubato.InitParams(tc.FVParamIndex, tc.Params, len(tc.Plaintext))

	runArenaBenchmark(b, "Rubato/HEKeyGen", func() {
		heRubato.HEKeyGen()
	})
	runArenaBenchmark(b, "Rubato/HalfBootKeyGen", func() {
		heRubato.HalfBootKeyGen()
	})

	heRubato.InitHalfBootstrapper()
	heRubato.InitEvaluator()
	heRubato.InitCoefficients()

	data := heRubato.RandomDataGen()
	nonces := heRubato.NonceGen()
	counter := make([]byte, 8)
	_, _ = rand.Read(counter)
	keyStream := make([][]uint64, heRubato.N)

	runArenaBenchmark(b, "Rubato/SymKeyStream", func() {
		for i := 0; i < heRubato.N; i++ {
			symRub := rubato.NewRubato(tc.Key, tc.Params)
			keyStream[i] = symRub.KeyStream(nonces[i], counter)
		}
	})

	heRubato.DataToCoefficients(data)
	runArenaBenchmark(b, "Rubato/EncryptSymData", func() {
		heRubato.EncodeEncrypt(keyStream)
	})

	heRubato.ScaleUp()
	runArenaBenchmark(b, "Rubato/EncSymKey", func() {
		_ = heRubato.InitFvRubato()
		heRubato.EncryptSymKey(tc.Key)
	})

	var fvKeyStreams []*RtF.Ciphertext
	runArenaBenchmark(b, "Rubato/FVKeyStream", func() {
		fvKeyStreams = heRubato.GetFvKeyStreams(nonces, counter)
	})

	heRubato.ScaleCiphertext(fvKeyStreams)
	runArenaBenchmark(b, "Rubato/HalfBoot", func() {
		_ = heRubato.HalfBoot()
	})
}
