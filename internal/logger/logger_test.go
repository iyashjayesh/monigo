package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	// Log should work without Init() being called.
	Log.Info("test message")
}

func TestInit(t *testing.T) {
	Init(slog.LevelWarn)
	// After Init, warn and above should work.
	Log.Warn("this should appear")
}

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer
	custom := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	SetLogger(custom)

	Log.Info("hello from custom logger")

	if !strings.Contains(buf.String(), "hello from custom logger") {
		t.Errorf("expected custom logger output, got %q", buf.String())
	}

	// Reset to default for other tests.
	Init(slog.LevelInfo)
}

func TestSetLoggerNilIgnored(t *testing.T) {
	before := Get()
	SetLogger(nil)
	after := Get()
	if before != after {
		t.Error("SetLogger(nil) should be a no-op")
	}
}

func TestGet(t *testing.T) {
	l := Get()
	if l == nil {
		t.Error("Get() should never return nil")
	}
}

func TestConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Log.Info("concurrent message")
		}()
	}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Init(slog.LevelDebug)
		}()
	}
	wg.Wait()
}
