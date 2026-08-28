package alerting

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

// recorder is a stub httpDoer that captures every delivery attempt.
type recorder struct {
	mu       sync.Mutex
	requests []AlertPayload
	status   int
	err      error
}

func (r *recorder) Do(req *http.Request) (*http.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.err != nil {
		return nil, r.err
	}

	var p AlertPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err == nil {
		r.requests = append(r.requests, p)
	}

	status := r.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Header:     make(http.Header),
	}, nil
}

func (r *recorder) calls() []AlertPayload {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]AlertPayload(nil), r.requests...)
}

// install swaps in a recorder and a rate limit for the duration of a test, and
// restores every piece of package state afterwards so tests do not leak into
// one another.
func install(t *testing.T, rl time.Duration) *recorder {
	t.Helper()
	rec := &recorder{}

	mu.Lock()
	prevClient, prevLimit, prevURL, prevLast := client, alertRateLimit, webhookURL, lastAlert
	client, alertRateLimit, lastAlert = rec, rl, time.Time{}
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		client, alertRateLimit, webhookURL, lastAlert = prevClient, prevLimit, prevURL, prevLast
		mu.Unlock()
	})
	return rec
}

// settle waits for the asynchronous delivery goroutine to reach the recorder.
func settle(t *testing.T, rec *recorder, want int) []AlertPayload {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := rec.calls(); len(got) >= want {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	return rec.calls()
}

func TestTriggerAlertDeliversPayload(t *testing.T) {
	rec := install(t, time.Minute)
	SetWebhookURL("https://example.com/webhook")

	TriggerAlert("test-service", 65.5, "Memory usage too high")

	got := settle(t, rec, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(got))
	}
	if got[0].ServiceName != "test-service" {
		t.Errorf("ServiceName = %q, want %q", got[0].ServiceName, "test-service")
	}
	if got[0].HealthScore != 65.5 {
		t.Errorf("HealthScore = %v, want 65.5", got[0].HealthScore)
	}
	if got[0].Message != "Memory usage too high" {
		t.Errorf("Message = %q, want %q", got[0].Message, "Memory usage too high")
	}
	if got[0].Timestamp.IsZero() {
		t.Error("expected a non-zero Timestamp")
	}
}

// An unconfigured webhook must produce no outbound traffic at all.
func TestTriggerAlertIsOptIn(t *testing.T) {
	rec := install(t, time.Minute)
	SetWebhookURL("")

	TriggerAlert("test-service", 10, "should not be delivered")

	time.Sleep(100 * time.Millisecond)
	if got := rec.calls(); len(got) != 0 {
		t.Errorf("expected no deliveries with no webhook configured, got %d", len(got))
	}
}

// A flapping service must not emit an alert on every collection tick, and the
// limiter is shared across system and service alerts.
func TestTriggerAlertIsRateLimited(t *testing.T) {
	rec := install(t, time.Hour)
	SetWebhookURL("https://example.com/webhook")

	TriggerAlert("test-service", 65.5, "first")
	settle(t, rec, 1)

	TriggerAlert("test-service", 60.0, "second")
	TriggerAlert("test-service", 55.0, "third")
	time.Sleep(100 * time.Millisecond)

	if got := rec.calls(); len(got) != 1 {
		t.Errorf("expected subsequent alerts to be rate limited, got %d deliveries", len(got))
	}
}

// Once the interval has elapsed, alerting resumes.
func TestTriggerAlertResumesAfterRateLimitWindow(t *testing.T) {
	rec := install(t, 50*time.Millisecond)
	SetWebhookURL("https://example.com/webhook")

	TriggerAlert("test-service", 65.5, "first")
	settle(t, rec, 1)

	time.Sleep(80 * time.Millisecond)
	TriggerAlert("test-service", 60.0, "second")

	if got := settle(t, rec, 2); len(got) != 2 {
		t.Errorf("expected delivery to resume after the window, got %d", len(got))
	}
}

// Delivery failures are logged, never fatal, and must not stall the caller.
func TestTriggerAlertToleratesDeliveryFailure(t *testing.T) {
	t.Run("non-2xx response", func(t *testing.T) {
		rec := install(t, time.Minute)
		rec.status = http.StatusInternalServerError
		SetWebhookURL("https://example.com/webhook")

		TriggerAlert("test-service", 20, "boom")
		if got := settle(t, rec, 1); len(got) != 1 {
			t.Errorf("expected the delivery to be attempted, got %d", len(got))
		}
	})

	t.Run("transport error", func(t *testing.T) {
		rec := install(t, time.Minute)
		rec.err = http.ErrHandlerTimeout
		SetWebhookURL("https://example.com/webhook")

		TriggerAlert("test-service", 20, "boom")
		time.Sleep(100 * time.Millisecond) // must not panic or block
	})
}

func TestWebhookURLRoundTrips(t *testing.T) {
	install(t, time.Minute)

	SetWebhookURL("https://example.com/hook")
	if got := GetWebhookURL(); got != "https://example.com/hook" {
		t.Errorf("GetWebhookURL() = %q, want %q", got, "https://example.com/hook")
	}
}

// TriggerAlert is called from the metrics collection path; concurrent calls must
// be safe. Run under -race.
func TestTriggerAlertConcurrent(t *testing.T) {
	install(t, time.Millisecond)
	SetWebhookURL("https://example.com/webhook")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); TriggerAlert("svc", 10, "concurrent") }()
		go func() { defer wg.Done(); _ = GetWebhookURL() }()
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)
}
