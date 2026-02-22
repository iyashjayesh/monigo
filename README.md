<p align="center">
  <img src="./static/assets/monigo-icon.png" width="200" title="Monigo Icon" alt="monigo-icon"/>
</p>

# MoniGo - Runtime Observability for Go Applications

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/monigo)](https://goreportcard.com/report/github.com/iyashjayesh/monigo)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/monigo?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/monigo)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub last commit](https://img.shields.io/github/last-commit/iyashjayesh/monigo)
![Tests](https://github.com/iyashjayesh/monigo/actions/workflows/test.yml/badge.svg)

**MoniGo** is a lightweight, embeddable observability library for Go services. It collects runtime metrics (CPU, memory, goroutines, disk/network I/O), traces function execution with pprof profiling, stores time-series data, and serves a real-time dashboard - all from a single `go get`.

<div align="center">
  <img src="monigo.gif" width="100%" alt="monigo-gif">
</div>

## Features

- **Function-Level Tracing** - Profile any function with CPU/memory pprof, adaptive sampling, and reflection-based argument capture
- **Pluggable Storage** - Persistent disk (tstorage) or volatile in-memory backends
- **Real-Time Dashboard** - Embedded web UI with system metrics, health scoring, goroutine inspection, and downloadable reports
- **Prometheus & OpenTelemetry** - Built-in `/metrics` endpoint and OTLP/gRPC export
- **Router Integration** - Works with `net/http`, Gin, Echo, Chi, Fiber, Gorilla Mux
- **Dashboard Security** - Basic Auth, API Key, IP Whitelist, Rate Limiting middleware
- **Headless Mode** - Run as a background telemetry agent without the dashboard
- **Builder API** - Type-safe, chainable configuration with validation

## Installation

```bash
go get github.com/iyashjayesh/monigo@latest
```

Requires **Go 1.22+**.

## Quick Start

```go
package main

import (
    "log"
    "math"
    "net/http"

    "github.com/iyashjayesh/monigo"
)

func main() {
    m := monigo.NewBuilder().
        WithServiceName("my-api").
        WithPort(8080).
        WithStorageType("memory").
        WithSamplingRate(100).
        Build()

    go func() {
        if err := m.Start(); err != nil {
            log.Fatalf("monigo: %v", err)
        }
    }()

    http.HandleFunc("/compute", func(w http.ResponseWriter, r *http.Request) {
        monigo.TraceFunction(r.Context(), heavyWork)
        w.Write([]byte("done"))
    })

    log.Fatal(http.ListenAndServe(":9000", nil))
}

func heavyWork() {
    var sum float64
    for i := 0; i < 1e8; i++ {
        sum += math.Sqrt(float64(i))
    }
}
```

Dashboard: `http://localhost:8080` - Your app: `http://localhost:9000`

## Configuration

All configuration is done via the builder pattern:

```go
m := monigo.NewBuilder().
    WithServiceName("order-service").       // Required
    WithPort(8080).                         // Dashboard port (default: 8080)
    WithStorageType("disk").                // "disk" or "memory" (default: "disk")
    WithRetentionPeriod("7d").              // Data retention (default: "7d")
    WithDataPointsSyncFrequency("5m").      // Metric flush interval (default: "5m")
    WithSamplingRate(100).                  // Trace 1 in N calls (default: 100)
    WithMaxCPUUsage(90).                    // Health threshold (default: 95%)
    WithMaxMemoryUsage(90).                 // Health threshold (default: 95%)
    WithMaxGoRoutines(500).                 // Health threshold (default: 100)
    WithHeadless(false).                    // true = no dashboard (default: false)
    WithTimeZone("UTC").                    // Timezone (default: "Local")
    WithLogLevel(slog.LevelInfo).           // Log level
    WithOTelEndpoint("localhost:4317").      // OTLP gRPC endpoint
    WithOTelHeaders(map[string]string{      // OTel auth headers
        "Authorization": "Bearer <token>",
    }).
    Build()
```

## Function Tracing

```go
// Simple function
monigo.TraceFunction(ctx, myFunc)

// Function with arguments
monigo.TraceFunctionWithArgs(ctx, processOrder, orderID, userID)

// Function with single return
result := monigo.TraceFunctionWithReturn(ctx, calculateTotal, items).(float64)

// Function with multiple returns
results := monigo.TraceFunctionWithReturns(ctx, validateInput, data)
val := results[0].(string)
err := results[1].(error)
```

Each traced call captures: execution time, memory delta, goroutine delta, and (at sampling rate) CPU/memory pprof profiles.

## Dashboard Security

```go
mw, stop := monigo.RateLimitMiddleware(100, time.Minute)
defer stop()

m := monigo.NewBuilder().
    WithServiceName("secure-api").
    WithPort(8080).
    WithDashboardMiddleware(
        monigo.BasicAuthMiddleware("admin", "s3cret"),
        mw,
    ).
    WithAPIMiddleware(
        monigo.APIKeyMiddleware("my-api-key"),
    ).
    Build()
```

## Router Integration

MoniGo integrates with any Go HTTP router:

```go
// Standard net/http
mux := http.NewServeMux()
monigo.RegisterDashboardHandlers(mux)

// Fiber
app.All("/monigo/*", adaptor.HTTPHandler(monigo.GetUnifiedHandler()))

// Gin / Echo / Chi - use GetAPIHandlers() map
for path, handler := range monigo.GetAPIHandlers() {
    router.GET(path, gin.WrapF(handler))
}
```

See [`example/router-integration/`](example/router-integration/) for complete examples.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/monigo/api/v1/metrics` | Current service statistics |
| GET | `/monigo/api/v1/service-info` | Service metadata |
| POST | `/monigo/api/v1/service-metrics` | Query time-series data |
| GET | `/monigo/api/v1/go-routines-stats` | Goroutine stack analysis |
| GET | `/monigo/api/v1/function` | Function trace summary |
| GET | `/monigo/api/v1/function-details` | pprof reports for a function |
| POST | `/monigo/api/v1/reports` | Aggregated report data |
| GET | `/metrics` | Prometheus scrape endpoint |

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Your Application                │
│                                                  │
│  monigo.TraceFunction(ctx, fn)                   │
│  monigo.Start() / monigo.Initialize()            │
└───────────┬─────────────────────┬────────────────┘
            │                     │
    ┌───────▼───────┐    ┌───────▼────────┐
    │   core/       │    │   monigo.go    │
    │  Metrics      │    │  Dashboard     │
    │  Collection   │    │  HTTP Server   │
    │  Health Score │    │  Middleware     │
    │  Profiling    │    │  Router Integ. │
    └───────┬───────┘    └───────┬────────┘
            │                     │
    ┌───────▼─────────────────────▼────────┐
    │           timeseries/                 │
    │  tstorage (disk) │ InMemoryStorage   │
    └───────┬──────────────────────────────┘
            │
    ┌───────▼───────────────────────────────┐
    │           exporters/                   │
    │  Prometheus Collector │ OTel Exporter  │
    └───────────────────────────────────────┘
            │
    ┌───────▼───────────────────────────────┐
    │           internal/                    │
    │  registry │ pipeline │ exporter │ log  │
    └───────────────────────────────────────┘
```

**Package responsibilities:**

| Package | Role |
|---------|------|
| `monigo` (root) | Public API, dashboard server, middleware, builder |
| `core` | System metric collection, function tracing, health scoring |
| `common` | Utilities, unit conversion, process info |
| `timeseries` | Storage abstraction (disk + in-memory) |
| `exporters` | Prometheus collector, OTel OTLP exporter |
| `internal/registry` | Thread-safe metric registry |
| `internal/pipeline` | Async metric export pipeline |
| `internal/exporter` | Exporter interface + fan-out |
| `internal/logger` | Race-safe structured logger (slog) |
| `models` | Shared data structures |
| `api` | HTTP handlers for all endpoints |

## Production Readiness Assessment

> This section is an honest evaluation of MoniGo's readiness for production use, written from the perspective of a distributed systems engineer.

### What Works Well

- **Zero-dependency embedding** - Single `go get`, no sidecar, no agent
- **Goroutine-safe metrics** - Atomic sampling, mutex-protected maps, `sync.Once` singletons
- **Adaptive profiling** - Sampling prevents pprof overhead in hot paths
- **Clean builder API** - Validation at build time catches misconfiguration early
- **Race-free logger** - `atomic.Pointer` eliminates the common `slog` data race

### Known Limitations & Risks

#### Security (Critical)

| Issue | Impact | Status |
|-------|--------|--------|
| `ViewFunctionMetrics` passes user input to `exec.Command("go", "tool", "pprof", ...)` | Command injection if `name` or `reportType` contains shell metacharacters | **Mitigated** - `exec.Command` does not invoke a shell, but `name` can reference arbitrary file paths |
| `X-Forwarded-For` trusted without proxy validation | IP whitelist bypass in `IPWhitelistMiddleware` | **Open** - Document: only use behind trusted reverse proxy |
| API key accepted via `?api_key=` query param | Key leaks in access logs, browser history, referrer headers | **Open** - Consider removing query param support |
| OTel defaults to insecure gRPC | Metrics sent unencrypted | **Open** - Default should be secure; require explicit opt-in for insecure |

#### Scalability

| Issue | Impact |
|-------|--------|
| `GetServiceStats()` calls gopsutil synchronously on every API request | At high dashboard QPS, CPU/memory reads contend on `/proc` |
| In-memory storage (`InMemoryStorage`) has no size bound | Unbounded growth if retention is not enforced |
| Time-series queries do linear scan (`O(n)`) in memory backend | Slow for large datasets |
| `functionMetrics` map capped at 10K entries with arbitrary eviction | Eviction is random (Go map iteration order), not LRU |

#### Reliability

| Issue | Impact |
|-------|--------|
| Storage singleton (`sync.Once`) cannot recover from init failure | If tstorage fails to open on startup, no retry is possible |
| `StoreServiceMetrics` failures are logged but not retried | Data points are silently lost |
| No circuit breaker on OTel export | If collector is down, export attempts continue indefinitely |
| `serviceInfo` and `retentionPeriod` in `common` package are unprotected global variables | Data race under concurrent `SetServiceInfo` / `GetServiceInfo` |

#### Testing

| Area | Coverage |
|------|----------|
| Core metrics, function tracing | Good |
| Registry, pipeline, exporter, logger | Good |
| Middleware (auth, rate limit, IP) | Good |
| API success paths | Partial - only error cases for POST endpoints |
| OTel exporter | None |
| Disk storage (`StorageWrapper`) | None |
| Dashboard serving, static files | None |
| Integration / E2E | None |

### Roadmap to Production Grade

#### A. Immediate Quick Wins

1. **Input validation on `ViewFunctionMetrics`** - Allowlist `reportType` to `["text", "top", "tree", "raw"]`
2. **Add `SECURITY.md`** - Document threat model, vulnerability reporting process
3. **Add `CONTRIBUTING.md`** - Code style, PR process, test requirements
4. **Add `golangci-lint` to CI** - Catch bugs the compiler misses
5. **Add coverage threshold** - Fail CI below 70%
6. **Protect `common.serviceInfo` with `sync.RWMutex`** - Prevent data race
7. **Document `X-Forwarded-For` trust requirement** - Must be behind trusted proxy

#### B. Medium-Term Refactors

1. **Decouple metric collection from HTTP handlers** - Cache `GetServiceStats()` result with a configurable TTL instead of computing on every request
2. **Add bounded buffer to in-memory storage** - Ring buffer or time-based eviction
3. **Make `Build()` return `(*Monigo, error)` instead of panicking** - Standard Go error handling
4. **Add environment variable support** - `MONIGO_PORT`, `MONIGO_SERVICE_NAME`, etc.
5. **Add retry with exponential backoff** - For OTel export and storage writes
6. **Integration test suite** - Start dashboard, hit API endpoints, verify responses
7. **OTel exporter tests** - Use an in-memory OTLP receiver
8. **Multi-version Go CI** - Test against Go 1.22, 1.23, 1.24

#### C. Long-Term Architecture Evolution

1. **Plugin architecture for storage** - Allow users to bring their own storage backend (e.g., ClickHouse, InfluxDB) via the `Storage` interface
2. **Distributed mode** - Aggregate metrics from multiple instances via a shared backend
3. **Config reloading** - Watch for config changes and apply without restart
4. **Structured events** - Emit lifecycle events (start, stop, health change) as structured logs or webhooks
5. **WASM dashboard** - Replace embedded HTML/JS with a lighter WASM-based UI
6. **Graduation path** - CONTRIBUTING.md, governance model, release automation, semantic versioning enforcement

## Examples

| Example | Path |
|---------|------|
| Basic usage | [`example/main.go`](example/main.go) |
| Function tracing | [`example/function-trace-example/`](example/function-trace-example/) |
| Gin integration | [`example/router-integration/gin-integration/`](example/router-integration/gin-integration/) |
| Echo integration | [`example/router-integration/echo-integration/`](example/router-integration/echo-integration/) |
| Fiber integration | [`example/router-integration/fiber-integration/`](example/router-integration/fiber-integration/) |
| Standard mux | [`example/router-integration/standard-mux-integration/`](example/router-integration/standard-mux-integration/) |
| Gorilla mux | [`example/router-integration/gorilla-mux-integration/`](example/router-integration/gorilla-mux-integration/) |
| Basic auth | [`example/security-examples/basic-auth/`](example/security-examples/basic-auth/) |
| API key auth | [`example/security-examples/api-key/`](example/security-examples/api-key/) |
| IP whitelist | [`example/security-examples/ip-whitelist-example/`](example/security-examples/ip-whitelist-example/) |
| Custom auth | [`example/security-examples/custom-auth/`](example/security-examples/custom-auth/) |

## Documentation

Full guides and API reference: **[iyashjayesh.github.io/monigo-website](https://iyashjayesh.github.io/monigo-website)**

| Topic | Link |
|-------|------|
| Introduction & Features | [Docs](https://iyashjayesh.github.io/monigo-website/guides/introduction/) |
| Configuration (Builder API) | [Docs](https://iyashjayesh.github.io/monigo-website/guides/configuration/) |
| Function Tracing | [Docs](https://iyashjayesh.github.io/monigo-website/guides/function-tracing/) |
| Router Integration | [Docs](https://iyashjayesh.github.io/monigo-website/guides/router-integration/) |
| Dashboard Security | [Docs](https://iyashjayesh.github.io/monigo-website/guides/security/) |
| Migration (v1 → v2) | [Docs](https://iyashjayesh.github.io/monigo-website/reference/migration-v1-to-v2/) |

## Contributing

We welcome contributions. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

If you find MoniGo useful, consider giving it a star.

## License

Apache 2.0 - see [LICENSE](LICENSE).

## Contact

For questions or feedback: [open an issue](https://github.com/iyashjayesh/monigo/issues) or reach out at `iyashjayesh@gmail.com` / [LinkedIn](https://www.linkedin.com/in/iyashjayesh/).

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iyashjayesh/monigo&type=Date)](https://star-history.com/#iyashjayesh/monigo&Date)
