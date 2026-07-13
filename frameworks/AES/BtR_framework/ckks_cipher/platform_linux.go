//go:build linux

package ckks_cipher

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func pinCurrentThread(cpu int) error {
	if cpu < 0 {
		return syscall.EINVAL
	}
	mask := make([]byte, cpu/8+1)
	mask[cpu/8] |= byte(1 << uint(cpu%8))
	_, _, errno := syscall.RawSyscall(syscall.SYS_SCHED_SETAFFINITY, 0, uintptr(len(mask)), uintptr(unsafe.Pointer(&mask[0])))
	if errno != 0 {
		return errno
	}
	return nil
}
func currentCPU() *int {
	var cpu uint32
	_, _, errno := syscall.RawSyscall(syscall.SYS_GETCPU, uintptr(unsafe.Pointer(&cpu)), 0, 0)
	if errno != 0 {
		return nil
	}
	value := int(cpu)
	return &value
}
func currentRSS() *int64 {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return nil
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return nil
	}
	pages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil
	}
	value := pages * int64(os.Getpagesize())
	return &value
}
func peakRSS() *int64 {
	var usage syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_SELF, &usage) != nil {
		return nil
	}
	value := usage.Maxrss * 1024
	return &value
}
