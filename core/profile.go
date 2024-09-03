package core

import (
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

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
	f.Close()
}

// WriteHeapProfile writes the current memory heap profile to the specified file.
func WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	runtime.GC() // Get up-to-date statistics
	return pprof.WriteHeapProfile(f)
}

// CollectGoRoutinesInfo returns the number of running Go routines and their stack traces split into separate goroutine blocks.
func CollectGoRoutinesInfo() models.GoRoutinesStatistic {
	// Creating a buffer to hold the stack trace
	stackBuffer := make([]byte, 1<<20)
	stackSize := runtime.Stack(stackBuffer, true)

	stackTrace := string(stackBuffer[:stackSize]) // converting the stack trace to a single string

	goroutineBlocks := SplitGoroutines(stackTrace)           // splitting the stack trace into separate goroutine blocks
	totalNumberOfRunningGoRoutines := runtime.NumGoroutine() // getting the total number of running goroutines

	return models.GoRoutinesStatistic{
		NumberOfGoroutines: totalNumberOfRunningGoRoutines,
		StackView:          goroutineBlocks,
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
