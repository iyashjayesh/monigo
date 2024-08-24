// Functions for CPU and memory profiling.

package monigo

import (
	"os"
	"runtime"
	"runtime/pprof"
)

func StartCPUProfile(filename string) (*os.File, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	pprof.StartCPUProfile(f)
	return f, nil
}

func StopCPUProfile(f *os.File) {
	pprof.StopCPUProfile()
	f.Close()
}

func WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	runtime.GC() // Get up-to-date statistics
	return pprof.WriteHeapProfile(f)
}
