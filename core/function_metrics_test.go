package core

import (
	"context"
	"testing"
)

func TestTraceFunction(t *testing.T) {
	SetSamplingRate(1) // Trace every call
	called := false
	TraceFunction(context.Background(), func() { called = true })

	if !called {
		t.Error("expected function to be called")
	}

	details := FunctionTraceDetails()
	if len(details) == 0 {
		t.Error("expected at least one function trace entry")
	}
}

func TestTraceFunctionWithArgs(t *testing.T) {
	SetSamplingRate(1)
	var got string
	fn := func(s string) { got = s }
	TraceFunctionWithArgs(context.Background(), fn, "hello")

	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestTraceFunctionWithArgs_WrongArgCount(t *testing.T) {
	SetSamplingRate(1)
	fn := func(a, b string) {}
	// Should not panic, just log and return
	TraceFunctionWithArgs(context.Background(), fn, "only-one")
}

func TestTraceFunctionWithArgs_NotAFunction(t *testing.T) {
	SetSamplingRate(1)
	// Should not panic when passed a non-function
	TraceFunctionWithArgs(context.Background(), "not-a-function")
}

func TestTraceFunctionWithReturn(t *testing.T) {
	SetSamplingRate(1)
	fn := func(a, b int) int { return a + b }
	result := TraceFunctionWithReturn(context.Background(), fn, 3, 4)

	if result.(int) != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestTraceFunctionWithReturns(t *testing.T) {
	SetSamplingRate(1)
	fn := func(s string) (string, int) { return s + "!", len(s) }
	results := TraceFunctionWithReturns(context.Background(), fn, "hi")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].(string) != "hi!" {
		t.Errorf("expected 'hi!', got %v", results[0])
	}
	if results[1].(int) != 2 {
		t.Errorf("expected 2, got %v", results[1])
	}
}

func TestSetSamplingRate(t *testing.T) {
	SetSamplingRate(1)
	if samplingRate.Load() != 1 {
		t.Errorf("expected sampling rate 1, got %d", samplingRate.Load())
	}

	// Rate < 1 should default to 1
	SetSamplingRate(0)
	if samplingRate.Load() != 1 {
		t.Errorf("expected sampling rate 1 for input 0, got %d", samplingRate.Load())
	}

	SetSamplingRate(-5)
	if samplingRate.Load() != 1 {
		t.Errorf("expected sampling rate 1 for negative input, got %d", samplingRate.Load())
	}

	SetSamplingRate(50)
	if samplingRate.Load() != 50 {
		t.Errorf("expected sampling rate 50, got %d", samplingRate.Load())
	}
}

func TestFunctionTraceDetailsReturnsCopy(t *testing.T) {
	SetSamplingRate(1)
	TraceFunction(context.Background(), func() {})

	details1 := FunctionTraceDetails()
	details2 := FunctionTraceDetails()

	// Mutating the returned map should not affect internal state
	for k := range details1 {
		delete(details1, k)
	}

	if len(details2) == 0 {
		t.Error("expected FunctionTraceDetails to return independent copies")
	}
}
