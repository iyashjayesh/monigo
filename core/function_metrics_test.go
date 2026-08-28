package core

import (
	"context"
	"strings"
	"testing"

	"github.com/iyashjayesh/monigo/models"
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

// ViewFunctionMetrics shells out to `go tool pprof` with a caller-supplied
// function name and report type. A name beginning with "-" is interpreted as a
// flag by pprof, and the report type is interpolated straight into the argument
// list, so both must be rejected before the exec.
func TestViewFunctionMetricsRejectsUnsafeArguments(t *testing.T) {
	metrics := &models.FunctionMetrics{
		CPUProfileFilePath: "testdata/does-not-exist.prof",
		MemProfileFilePath: "testdata/does-not-exist.prof",
	}

	unsafeNames := []string{
		"-output=/tmp/pwned",
		"-http=:8080",
		"name; rm -rf /",
		"name && whoami",
		"name | cat",
		"name $(id)",
		"name `id`",
		"name\nwhoami",
	}

	for _, name := range unsafeNames {
		got := ViewFunctionMetrics(name, "text", metrics)
		if !strings.Contains(got.FunctionCodeTrace, "Invalid function name") {
			t.Errorf("name %q: expected the code trace to be rejected, got %q", name, got.FunctionCodeTrace)
		}
		if !strings.Contains(got.CoreProfile.CPU, "Invalid function name") {
			t.Errorf("name %q: expected the CPU profile to be rejected, got %q", name, got.CoreProfile.CPU)
		}
	}
}

func TestViewFunctionMetricsRejectsUnknownReportType(t *testing.T) {
	metrics := &models.FunctionMetrics{
		CPUProfileFilePath: "testdata/does-not-exist.prof",
		MemProfileFilePath: "testdata/does-not-exist.prof",
	}

	for _, reportType := range []string{"", "svg", "-http=:8080", "web"} {
		got := ViewFunctionMetrics("SafeName", reportType, metrics)
		if !strings.Contains(got.CoreProfile.CPU, "Invalid report type") {
			t.Errorf("reportType %q: expected rejection, got %q", reportType, got.CoreProfile.CPU)
		}
	}
}

func TestViewFunctionMetricsAcceptsSafeArguments(t *testing.T) {
	metrics := &models.FunctionMetrics{
		CPUProfileFilePath: "testdata/does-not-exist.prof",
		MemProfileFilePath: "testdata/does-not-exist.prof",
	}

	for _, reportType := range []string{"text", "traces", "tree"} {
		got := ViewFunctionMetrics("main.SafeName", reportType, metrics)
		if strings.Contains(got.CoreProfile.CPU, "Invalid report type") {
			t.Errorf("reportType %q was rejected but should be allowed", reportType)
		}
		if strings.Contains(got.CoreProfile.CPU, "Invalid function name") {
			t.Errorf("reportType %q: safe name was rejected", reportType)
		}
	}
}

// A nil argument used to panic: reflect.ValueOf(nil) yields an invalid Value and
// calling .Type() on it panics. Nillable parameter kinds must accept nil as the
// zero value; non-nillable kinds must be rejected with a log, not a panic.
func TestTraceFunctionWithArgs_NilArguments(t *testing.T) {
	SetSamplingRate(1)

	t.Run("nil into a pointer parameter", func(t *testing.T) {
		got := "not called"
		TraceFunctionWithArgs(context.Background(), func(p *int) {
			if p == nil {
				got = "nil"
			}
		}, nil)
		if got != "nil" {
			t.Errorf("expected the function to run with a nil pointer, got %q", got)
		}
	})

	t.Run("nil into an interface parameter", func(t *testing.T) {
		called := false
		TraceFunctionWithArgs(context.Background(), func(v interface{}) {
			called = true
		}, nil)
		if !called {
			t.Error("expected the function to run with a nil interface")
		}
	})

	t.Run("nil into a slice and map parameter", func(t *testing.T) {
		called := false
		TraceFunctionWithArgs(context.Background(), func(s []string, m map[string]int) {
			called = s == nil && m == nil
		}, nil, nil)
		if !called {
			t.Error("expected the function to run with nil slice and map")
		}
	})

	t.Run("nil into a non-nillable parameter is rejected", func(t *testing.T) {
		called := false
		TraceFunctionWithArgs(context.Background(), func(n int) { called = true }, nil)
		if called {
			t.Error("expected nil to be rejected for an int parameter")
		}
	})
}

func TestTraceFunctionWithReturns_NilArguments(t *testing.T) {
	SetSamplingRate(1)

	results := TraceFunctionWithReturns(context.Background(), func(p *int) string {
		if p == nil {
			return "nil"
		}
		return "non-nil"
	}, nil)

	if len(results) != 1 {
		t.Fatalf("expected 1 return value, got %d", len(results))
	}
	if results[0] != "nil" {
		t.Errorf("expected %q, got %v", "nil", results[0])
	}
}
