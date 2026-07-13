//go:build !amd64

package ckks_cipher

func RDTSCPAvailable() bool { return false }
func Rdtsc() uint64         { return 0 }
func tscStart() uint64      { return 0 }
func tscEnd() uint64        { return 0 }
