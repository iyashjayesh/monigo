# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Type and density follow the design's scale.** Metric tiles are now an uppercase micro-label over one large tabular number, card titles and body text sit on a six-step scale topping out at 24px rather than the vendored template's defaults, card padding is 16px, and table rows are tighter with a sticky header. Every `font-size` and `border-radius` in the project stylesheet now comes from a token: 12 ad-hoc sizes and 5 ad-hoc radii reduced to 6 and 3
- Metric labels drop their trailing colons, since an uppercase eyebrow does not take one
- **The dashboard adopts the new design's palette and collapses the dark theme onto it.** Colours now come from the design system: a cyan accent (`#00B4E6` dark / `#00758F` light) replacing the orange, five surface planes per theme, and five semantic states. Hex literals in real declarations drop from 105 to 2; `body.dark-theme` component rules from 53 to 3. Every text pair clears WCAG AA and every control boundary clears 3:1, verified in a running dashboard in both themes
- Focus rings restored. The vendored `.btn:focus { outline: none; box-shadow: none }` stripped every focus indicator with no replacement, so tabbing the dashboard moved nothing visually
- Stack traces and pprof output render in IBM Plex Mono at a fixed size on a recessed surface. The previous rule asked for `'Fira Code'`, which is not shipped and never resolved
- Table numerics use `tabular-nums`, so live columns stop jittering on each poll

### Changed
- Cards, tables and muted text now read their colours from the design tokens rather than hardcoded hex literals. Colours shift slightly onto the corrected palette and every affected text pair gains contrast -- dark muted text goes from 5.78:1 to 8.38:1, dark card text from 13.34:1 to 14.98:1

### Added
- Design token layer in `static/css/monigo-styles.css`: 122 custom properties covering type, spacing, radius, elevation, five surface planes, four ink tones, five semantic states and an eight-colour chart series palette, in matched light and dark sets. Purely additive -- nothing consumes them yet, and removing both rules from the live stylesheet changes the computed style of zero of 713 sampled elements. Every documented contrast ratio is verified: text pairs clear 4.5:1, control borders and the focus ring clear 3:1

### Changed
- **The embedded dashboard shrank from 17.1 MB to 7.9 MB (-54%), and from 71 files to 46.** Everything under `static/` is compiled into the consuming service's binary by `//go:embed static/*` and downloaded by every `go get`, so this is weight every user carried whether or not they ever opened the dashboard. Removed: `assets/ss/d1-d10.png` (7.6 MB), Product Hunt marketing images, an unused animated logo and dropdown arrow, and a favicon variant set -- 28 files, none referenced by any page, stylesheet, script or document. Note this does not shrink existing clones, since the blobs remain in git history; the module zip is what `go get` downloads

### Fixed
- **The dashboard no longer requires internet access.** Font Awesome 4.7.0 and html2canvas were loaded from `cdnjs`, and the vendored template CSS pulled a Lato webfont from Google Fonts, so on an airgapped host -- where a lot of production Go services run -- icons were blank boxes and the screenshot feature was dead. All three fetches are gone: icons are an inline SVG sprite, html2canvas is vendored, and the font imports are removed
- **Mobile navigation was impossible in every network condition.** The sidebar and navbar menu buttons were `<i>` elements carrying `las la-bars` and `ri-menu-*` classes -- Line Awesome and Remix Icon -- and neither font was ever loaded: `font-family: remixicon` appeared in the vendored CSS with no `@font-face` rule and no font file anywhere in `static/`. They rendered as zero-size empty elements, so the sidebar could not be opened below 1300px and the navbar below 992px. Both are now real 22px controls
- Removed `static/css/core/intro.css` (192 KB), referenced by no page and carrying a third Google Fonts import
- **Memory Distribution pie plotted values of different magnitudes against each other.** The chart parsed a number out of the pre-formatted display strings (`"2.68 MB"`, `"12.93 GB"`), which discards the unit -- so a service using 2.68 MB rendered as 14.3% of the pie instead of 0.02%, a ~700x overstatement of its footprint
- **Heap Memory Usage chart mixed units under an "MB" axis.** It read `mem_stats_records[].record_value`, a display number whose unit lives in a separate `record_unit` field that the chart ignored. Once heap use crosses 1 GB, `HeapSys` reports in GB while `HeapAlloc` is still in MB, and the taller bar renders shorter. It now reads `raw_mem_stats_records`, which is in one consistent unit
- **CPU Statistics pie plotted `total_cores` as a slice alongside its own components**, so the chart always summed to twice the real core count. The third slice is now idle cores

### Added
- Unformatted memory values on the metrics API: `total_system_memory_bytes`, `memory_used_by_system_bytes`, `memory_used_by_service_bytes`, `available_memory_bytes`, `stack_memory_usage_bytes`, `gc_pause_duration_ms`. Additive -- the existing formatted string fields are unchanged. Anything doing arithmetic on or plotting these metrics must use the raw fields, since the strings carry a unit suffix

## [1.3.0] - 2026-08-28

### Security
- `BasicAuthMiddleware` and `APIKeyMiddleware` now compare credentials with `crypto/subtle.ConstantTimeCompare` -- `!=` short-circuits on the first differing byte and leaks length and prefix information to a timing oracle
- `ViewFunctionMetrics` validates the function name and report type before invoking `go tool pprof` -- a name beginning with `-` was interpreted as a flag by pprof, and the report type was interpolated straight into the argument list
- `BasicAuthMiddleware` now sends a `WWW-Authenticate` challenge when a request carries no credentials at all

### Fixed
- **Silently truncated goroutine dumps**: `CollectGoRoutinesInfo` used a fixed 1 MB buffer with a single `runtime.Stack` call. `runtime.Stack` truncates without reporting it, so an application with thousands of goroutines got a clipped trace and undercounted -- precisely the situation goroutine leak detection needs to be accurate in. It now grows the buffer and retries, capped at 64 MB
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
- Goroutine leak detection: every collection cycle evaluates all goroutine stacks for stale goroutines (blocked past a threshold, default 24h, configurable via `WithStaleGoroutineThreshold`) and for call stacks growing monotonically across the last 5 cycles. The verdict is carried on `GoRoutinesStatistic.LeakReport` and `ServiceStats.GoroutineLeakReport`, surfaced as a warning panel on the Go Routines Stats page, and raises the alert webhook when configured
- `core.AnalyzeGoroutineLeaks()`, `core.SetStaleGoroutineThreshold()`, `core.GetStaleGoroutineThreshold()`, and `core.ResetLeakDetectionState()`
- Health breach webhook alerts via `WithAlertWebhook(url)` -- an opt-in `POST` fired when the system or service health score drops below 70. Dispatch is off the collection path, requests time out after 10 seconds, and deliveries are rate limited to one per 5 minutes across both alert types. `Build()` rejects a URL that is not `http://` or `https://`
- `common.GetServiceName()` -- accessor for the configured service name
- `core.GetThresholds()` -- thread-safe accessor for the configured health thresholds
- `Build()` validates `MaxCPUUsage` and `MaxMemoryUsage` are within 0-100 and `MaxGoRoutines` is non-negative
- Regression tests for every fix above

### Changed
- `monigo.gif` re-encoded from 65 MB to 24 MB -- the module zip is downloaded on every `go get`
- Argument marshalling shared between `TraceFunctionWithArgs` and `TraceFunctionWithReturns` instead of duplicated

## [2.0.0] - 2026-02-10

> **Never published.** This release was tagged `v2.0.0`, but the Go module proxy
> rejects it: `go.mod` declares `module github.com/iyashjayesh/monigo` with no
> `/v2` suffix, and Go requires the major version to appear in the module path
> from v2 onwards. `go get` therefore continued to serve `v1.2.0`, and everything
> below was unreachable to anyone installing the library.
>
> ```
> $ curl -s https://proxy.golang.org/github.com/iyashjayesh/monigo/@v/v2.0.0.info
> not found: invalid version: module contains a go.mod file, so module path must
> match major version ("github.com/iyashjayesh/monigo/v2")
> ```
>
> The release line was renumbered rather than renaming the module: `/v2` would be
> permanent for an API that is not v2-shaped, and per the proxy there were no v2
> consumers to preserve. The changes below shipped in `1.3.0`.

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
