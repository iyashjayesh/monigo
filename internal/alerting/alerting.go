package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/internal/logger"
)

// httpDoer is the subset of *http.Client that alert delivery needs. It exists so
// tests can substitute a transport without mutating http.DefaultClient.
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	// deliveryTimeout bounds a single webhook request so a hanging endpoint
	// cannot pin a goroutine indefinitely.
	deliveryTimeout = 10 * time.Second
	// defaultRateLimit is the minimum interval between deliveries. A flapping
	// service would otherwise emit an alert on every collection tick.
	defaultRateLimit = 5 * time.Minute
)

var (
	mu             sync.RWMutex
	webhookURL     string
	lastAlert      time.Time
	alertRateLimit          = defaultRateLimit
	client         httpDoer = &http.Client{Timeout: deliveryTimeout}
)

// SetWebhookURL sets the webhook URL.
func SetWebhookURL(url string) {
	mu.Lock()
	defer mu.Unlock()
	webhookURL = url
}

// GetWebhookURL returns the webhook URL.
func GetWebhookURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return webhookURL
}

// currentClient returns the client used to deliver alerts.
func currentClient() httpDoer {
	mu.RLock()
	defer mu.RUnlock()
	return client
}

// AlertPayload defines the schema for the webhook payload.
type AlertPayload struct {
	ServiceName string    `json:"service_name"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
	HealthScore float64   `json:"health_score"`
}

// TriggerAlert asynchronously dispatches a health breach webhook if configured.
func TriggerAlert(serviceName string, score float64, message string) {
	mu.Lock()
	url := webhookURL
	if url == "" {
		mu.Unlock()
		return
	}
	if time.Since(lastAlert) < alertRateLimit {
		mu.Unlock()
		return
	}
	lastAlert = time.Now()
	mu.Unlock()

	go func() {
		payload := AlertPayload{
			ServiceName: serviceName,
			Timestamp:   time.Now(),
			Message:     message,
			HealthScore: score,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			logger.Log.Error("failed to marshal alert payload", "error", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
		if err != nil {
			logger.Log.Error("failed to create alert request", "error", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := currentClient().Do(req)
		if err != nil {
			logger.Log.Error("failed to send alert webhook", "error", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			logger.Log.Error("alert webhook returned non-success status", "status", resp.Status)
		}
	}()
}
