package main

import (
	"fmt"
	"runtime"

	benchmark "hhe-arena/benchmark/utils"
)

func main() {
	_ = benchmark.PolicyWarn
	fmt.Printf("hhe-go-probe go=%s os=%s arch=%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
