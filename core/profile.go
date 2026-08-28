package core

import (
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/models"
)

// StartCPUProfile starts the CPU profile and writes it to the specified file.
func StartCPUProfile(filename string) (*os.File, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	pprof.StartCPUProfile(f)
	return f, nil
}

// StopCPUProfile stops the current CPU profile and writes it to the specified file.
func StopCPUProfile(f *os.File) {
	pprof.StopCPUProfile()
	if f != nil {
		f.Close()
	}
}

// WriteHeapProfile writes the current memory heap profile to the specified file.
func WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	runtime.GC() // Get up-to-date statistics
	return pprof.WriteHeapProfile(f)
}

// maxStackBufferSize caps the buffer used to capture all goroutine stacks. An
// application with tens of thousands of goroutines produces a very large dump;
// beyond this we accept truncation rather than allocating without bound.
const maxStackBufferSize = 64 << 20

// initialStackBufferSize is the first buffer tried before any doubling.
const initialStackBufferSize = 1 << 20

// captureAllStacks returns the full stack trace for every goroutine.
//
// runtime.Stack truncates silently when the destination buffer is too small and
// reports only how many bytes it wrote, so a single fixed-size call cannot tell
// a complete dump from a clipped one. It is retried with a larger buffer while
// the result exactly fills the buffer -- the only observable signal of
// truncation -- which matters most for an application leaking goroutines, where
// a clipped trace would undercount precisely when the data is needed.
func captureAllStacks() string {
	return captureStacks(runtime.Stack, initialStackBufferSize, maxStackBufferSize)
}

// captureStacks holds the retry logic, with the dump function injected so the
// grow-and-retry path can be tested without needing a process that genuinely
// has a multi-megabyte stack dump.
func captureStacks(dump func(buf []byte, all bool) int, initialSize, maxSize int) string {
	size := initialSize
	for {
		buf := make([]byte, size)
		n := dump(buf, true)
		if n < len(buf) {
			return string(buf[:n])
		}
		if size >= maxSize {
			logger.Log.Warn("goroutine stack dump truncated", "buffer_bytes", size)
			return string(buf[:n])
		}
		size *= 2
	}
}

// CollectGoRoutinesInfo returns the number of running Go routines and their stack traces split into separate goroutine blocks.
func CollectGoRoutinesInfo() models.GoRoutinesStatistic {
	stackTrace := captureAllStacks()

	goroutineBlocks := SplitGoroutines(stackTrace)           // splitting the stack trace into separate goroutine blocks
	totalNumberOfRunningGoRoutines := runtime.NumGoroutine() // getting the total number of running goroutines

	return models.GoRoutinesStatistic{
		NumberOfGoroutines: totalNumberOfRunningGoRoutines,
		StackView:          goroutineBlocks,
		LeakReport:         leakReportFor(goroutineBlocks),
	}
}

// SplitGoroutines splits the input stack trace into separate goroutine blocks based on new lines and "goroutine" identifiers.
func SplitGoroutines(stackTrace string) []string {
	var goroutines []string
	var currentGoroutine strings.Builder

	lines := strings.Split(stackTrace, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "goroutine ") {
			if currentGoroutine.Len() > 0 {
				goroutines = append(goroutines, currentGoroutine.String())
				currentGoroutine.Reset()
			}
		}
		currentGoroutine.WriteString(line + "\n")
	}

	// Appening the last goroutine block if there's any content
	if currentGoroutine.Len() > 0 {
		goroutines = append(goroutines, currentGoroutine.String())
	}

	return goroutines
}
