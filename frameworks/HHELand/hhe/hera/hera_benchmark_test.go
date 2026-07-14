package hera

import (
	"HHELand/rtf_integration"
	"HHELand/sym/hera"
	"HHELand/utils"
	"fmt"
	"sync"
	"testing"
)

func BenchmarkHera(b *testing.B) {
	for _, tc := range hera.TestVector {
		if tc.Params.Rounds == 4 {
			continue
		}
		benchHEHera(tc, b)
	}
}

func runArenaBenchmark(b *testing.B, operation string, target func()) {
	var once sync.Once
	b.Run(operation, func(b *testing.B) {
		once.Do(func() { utils.BenchmarkIter(operation, 1, target) })
		b.ReportMetric(0, "ns/op")
	})
}

func benchHEHera(tc hera.TestContext, b *testing.B) {
	fmt.Println(testString("HERA", tc.Params))
	if testing.Short() {
		b.Skip("skipping benchmark in short mode.")
	}

	logger := utils.NewLogger(utils.DEBUG)
	logger.PrintDataLen(tc.Key)

	heHera := NewHEHera()
	var data [][]float64
	var nonces [][]byte
	var keyStream [][]uint64

	heHera.InitParams(tc.FVParamIndex, tc.Params)

	runArenaBenchmark(b, "HERA/HEKeyGen", func() {
		heHera.HEKeyGen()
	})
	runArenaBenchmark(b, "HERA/HalfBootKeyGen", func() {
		heHera.HalfBootKeyGen(tc.Radix)
	})

	heHera.InitHalfBootstrapper()
	heHera.InitEvaluator()
	heHera.InitCoefficients()

	if heHera.fullCoefficients {
		data = heHera.RandomDataGen(heHera.params.N())
		nonces = heHera.NonceGen(heHera.params.N())
		keyStream = make([][]uint64, heHera.params.N())
		runArenaBenchmark(b, "HERA/SymKeyStream", func() {
			for i := 0; i < heHera.params.N(); i++ {
				symHera := hera.NewHera(tc.Key, tc.Params)
				keyStream[i] = symHera.KeyStream(nonces[i])
			}
		})
		heHera.DataToCoefficients(data, heHera.params.N())
		runArenaBenchmark(b, "HERA/EncryptSymData", func() {
			heHera.EncodeEncrypt(keyStream, heHera.params.N())
		})
	} else {
		data = heHera.RandomDataGen(heHera.params.Slots())
		nonces = heHera.NonceGen(heHera.params.Slots())
		keyStream = make([][]uint64, heHera.params.Slots())
		runArenaBenchmark(b, "HERA/SymKeyStream", func() {
			for i := 0; i < heHera.params.Slots(); i++ {
				symHera := hera.NewHera(tc.Key, tc.Params)
				keyStream[i] = symHera.KeyStream(nonces[i])
			}
		})
		heHera.DataToCoefficients(data, heHera.params.Slots())
		runArenaBenchmark(b, "HERA/EncryptSymData", func() {
			heHera.EncodeEncrypt(keyStream, heHera.params.Slots())
		})
	}

	heHera.ScaleUp()

	runArenaBenchmark(b, "HERA/EncSymKey", func() {
		_ = heHera.InitFvHera()
		heHera.EncryptSymKey(tc.Key)
	})

	var fvKeyStreams []*RtF.Ciphertext
	runArenaBenchmark(b, "HERA/FVKeyStream", func() {
		fvKeyStreams = heHera.GetFvKeyStreams(nonces)
	})

	heHera.ScaleCiphertext(fvKeyStreams)
	runArenaBenchmark(b, "HERA/HalfBoot", func() {
		_ = heHera.HalfBoot()
	})
}
