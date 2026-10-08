package he

import (
	"os"
	"testing"

	"github.com/tuneinsight/lattigo/v6/examples"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func TestPureBgv(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in -short mode")
	}
	TestVector := []bgv.ParametersLiteral{
		examples.BGVParamsN13QP218,
		examples.BGVParamsN14QP438,
		examples.BGVParamsN15QP880,
	}
	switch os.Getenv("HHE_BENCH_VARIANT") {
	case "":
		PureBgv(TestVector[1])
	case "hheland-bgv-n13qp218":
		PureBgv(TestVector[0])
	case "hheland-bgv-n14qp438":
		PureBgv(TestVector[1])
	case "hheland-bgv-n15qp880":
		PureBgv(TestVector[2])
	case "hheland-ckks-n14qp438", "hheland-ckks-n15qp881", "hheland-ckks-pn16qp1761":
		t.Skip("selected matrix variant is CKKS")
	default:
		t.Fatalf("unsupported HHE_BENCH_VARIANT %q for BGV", os.Getenv("HHE_BENCH_VARIANT"))
	}
}
