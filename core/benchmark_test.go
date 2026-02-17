package core

import (
	"context"
	"runtime"
	"testing"
)

func BenchmarkGetServiceStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetServiceStats(context.Background())
	}
}

func BenchmarkGetCoreStatistics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetCoreStatistics()
	}
}

func BenchmarkGetMemoryStatistics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMemoryStatistics()
	}
}

func BenchmarkCalculateOverallLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateOverallLoad(45.5, 60.2)
	}
}

func BenchmarkConstructRawMemStats(b *testing.B) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConstructRawMemStats(m)
	}
}

func BenchmarkTraceFunction(b *testing.B) {
	SetSamplingRate(1000) // Low sampling to measure dispatch overhead
	f := func() {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TraceFunction(context.Background(), f)
	}
}

func BenchmarkTraceFunctionWithArgs(b *testing.B) {
	SetSamplingRate(1000)
	f := func(a int, s string) {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TraceFunctionWithArgs(context.Background(), f, 42, "test")
	}
}
