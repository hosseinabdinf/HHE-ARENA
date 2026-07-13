package ckks_cipher

//go:noescape
func cpuid(eax, ecx uint32) (a, b, c, d uint32)

//go:noescape
func tscStart() uint64

//go:noescape
func tscEnd() uint64

func RDTSCPAvailable() bool {
	maximum, _, _, _ := cpuid(0x80000000, 0)
	if maximum < 0x80000001 {
		return false
	}
	_, _, _, features := cpuid(0x80000001, 0)
	return features&(1<<27) != 0
}

func Rdtsc() uint64 { return tscStart() }
