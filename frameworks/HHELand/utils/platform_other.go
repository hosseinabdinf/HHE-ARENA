//go:build !linux

package utils

import "errors"

func pinCurrentThread(cpu int) error { return errors.New("CPU affinity requires Linux") }
func currentCPU() *int               { return nil }
func currentRSS() *int64             { return nil }
func peakRSS() *int64                { return nil }
