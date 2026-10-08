package he

import (
	"os"
	"testing"

	"github.com/tuneinsight/lattigo/v6/examples"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func TestPureCKKS(t *testing.T) {
	TestVector := []ckks.ParametersLiteral{
		examples.CKKSRealParamsN14QP438,
		examples.CKKSRealParamsN15QP881,
		examples.CKKSRealParamsPN16QP1761,
	}
	selected := os.Getenv("HHE_BENCH_VARIANT")
	if selected == "" {
		for _, tc := range TestVector {
			PureCKKS(tc)
		}
		return
	}
	switch selected {
	case "hheland-ckks-n14qp438":
		PureCKKS(TestVector[0])
	case "hheland-ckks-n15qp881":
		PureCKKS(TestVector[1])
	case "hheland-ckks-pn16qp1761":
		PureCKKS(TestVector[2])
	case "hheland-bgv-n13qp218", "hheland-bgv-n14qp438", "hheland-bgv-n15qp880":
		t.Skip("selected matrix variant is BGV")
	default:
		t.Fatalf("unsupported HHE_BENCH_VARIANT %q for CKKS", selected)
	}
}

func TestChebyshevCKKS(t *testing.T) {
	TestVector := []ckks.ParametersLiteral{
		examples.CKKSRealParamsN14QP438,
		examples.CKKSRealParamsN15QP881,
		examples.CKKSRealParamsPN16QP1761,
	}
	selected := os.Getenv("HHE_BENCH_VARIANT")
	if selected == "" {
		for _, tc := range TestVector {
			ChebyshevCKKS(tc)
		}
		return
	}
	switch selected {
	case "hheland-ckks-n14qp438":
		ChebyshevCKKS(TestVector[0])
	case "hheland-ckks-n15qp881":
		ChebyshevCKKS(TestVector[1])
	case "hheland-ckks-pn16qp1761":
		ChebyshevCKKS(TestVector[2])
	case "hheland-bgv-n13qp218", "hheland-bgv-n14qp438", "hheland-bgv-n15qp880":
		t.Skip("selected matrix variant is BGV")
	default:
		t.Fatalf("unsupported HHE_BENCH_VARIANT %q for Chebyshev CKKS", selected)
	}
}
