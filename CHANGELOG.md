# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Security
- `BasicAuthMiddleware` and `APIKeyMiddleware` now compare credentials with `crypto/subtle.ConstantTimeCompare` -- `!=` short-circuits on the first differing byte and leaks length and prefix information to a timing oracle
- `ViewFunctionMetrics` validates the function name and report type before invoking `go tool pprof` -- a name beginning with `-` was interpreted as a flag by pprof, and the report type was interpolated straight into the argument list
- `BasicAuthMiddleware` now sends a `WWW-Authenticate` challenge when a request carries no credentials at all

### Fixed
- **Inverted health score**: `calculateServiceHealth` clamped the score to `100` -- the best possible value -- when usage exceeded every threshold, so a fully degraded service reported as healthy. It now returns `0`, matching `calculateSystemHealth`
- **Unbounded memory growth**: `InMemoryStorage.InsertRows` ignored the retention period and never evicted, so `WithStorageType("memory")` grew for the life of the process
- **`WithStorageType` had no effect**: the storage singleton was initialised by `SetDataPointsSyncFrequency` before `SetStorageType` was applied, so `"memory"` silently fell through to disk
- **Data race** on `serviceHealthThresholds` between `ConfigureServiceThresholds` and the metrics collection goroutine
- **Panic** in `common.ConvertToMB` on memory strings shorter than three characters, reachable from health calculation; `calculateMemoryUsagePercentage` also guards against a zero total
- **Panic** in `TraceFunctionWithArgs` / `TraceFunctionWithReturns` when passed a `nil` argument -- `reflect.ValueOf(nil).Type()` panics. Nillable parameter kinds now accept `nil` as the zero value
- `CloseStorage()` was permanently terminal via `sync.Once`; a subsequent `GetStorageInstance()` returned the closed handle instead of re-initialising
- Goroutine leak: each call to `SetDataPointsSyncFrequency` started a ticker goroutine without stopping the previous one
- Goroutine leak: `RateLimitMiddleware`'s cleanup goroutine is now stopped by `Shutdown()`, not only by the caller invoking the returned `stop`
- File descriptor leak in `WriteHeapProfile`; nil check in `StopCPUProfile`
- `registry.GetAll()` documented a snapshot copy but shared the `Labels` map with the live registry

### Added
- Health breach webhook alerts via `WithAlertWebhook(url)` -- an opt-in `POST` fired when the system or service health score drops below 70. Dispatch is off the collection path, requests time out after 10 seconds, and deliveries are rate limited to one per 5 minutes across both alert types. `Build()` rejects a URL that is not `http://` or `https://`
- `common.GetServiceName()` -- accessor for the configured service name
- `core.GetThresholds()` -- thread-safe accessor for the configured health thresholds
- `Build()` validates `MaxCPUUsage` and `MaxMemoryUsage` are within 0-100 and `MaxGoRoutines` is non-negative
- Regression tests for every fix above

### Changed
- `monigo.gif` re-encoded from 65 MB to 24 MB -- the module zip is downloaded on every `go get`
- Argument marshalling shared between `TraceFunctionWithArgs` and `TraceFunctionWithReturns` instead of duplicated

## [2.0.0] - 2026-02-10

### Breaking Changes
- Renamed `GetRuningPort()` to `GetRunningPort()`
- Renamed `ViewFunctionMaetrtics()` to `ViewFunctionMetrics()` (api package)
- `Build()` now panics on invalid config (missing ServiceName, bad port, bad StorageType)
- All API endpoints now enforce HTTP methods (GET/POST) -- wrong method returns 405
- `isStaticFile()` no longer bypasses auth for `.html` files
- `context.Context` added as first parameter to `TraceFunction`, `TraceFunctionWithArgs`, `TraceFunctionWithReturn`, `TraceFunctionWithReturns`, `GetServiceStats`
- Structured logging via `log/slog` replaces `log.Printf` -- use `WithLogger()` / `WithLogLevel()` to customize

### Fixed
- **Data loss**: removed `PurgeStorage()` from startup -- historical data now survives restarts
- **Data race** on `samplingRate` -- now uses `sync/atomic`
- `FunctionTraceDetails()` returns deep copy instead of raw map pointer
- Replaced `http.DefaultServeMux` with dedicated mux (prevents route collisions)
- All API handlers check marshal errors and return proper 500s
- Hardcoded `host=server1` label replaced with `os.Hostname()`
- Duplicate `"sys"` key in raw memory stats
- `GCCPUFraction` and counter metrics no longer incorrectly converted as bytes
- `ConvertBytesToUnit` now uses base-1024 (was base-1000)
- Prometheus exporter uses raw float64 values instead of parsing formatted strings
- Replaced deprecated `ioutil.ReadAll` with `io.ReadAll`

### Added
- Graceful shutdown with SIGINT/SIGTERM handling
- Builder validation at `Build()` time
- Comprehensive test suite across all packages (core, api, common, timeseries, config)
- Benchmarks for hot paths (core, timeseries, common)
- OpenTelemetry exporter option via `WithOTelEndpoint()`
- Structured logging via `log/slog` with `WithLogger()` and `WithLogLevel()` builder options
- `context.Context` propagation through public API
- Decoupled `Storage` interface from tstorage types (monigo-owned `Label`, `DataPoint`, `Row`)

### Changed
- CI updated to Go 1.24, with race detector and `go vet`
- Storage interface uses monigo-owned types instead of tstorage types
