# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
