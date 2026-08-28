# MoniGo Development Plan

PR-level execution plan. Companion to [ROADMAP.md](ROADMAP.md), which carries the
*why* and the evidence behind each item; this file carries the *what ships in
which PR*.

Last updated 2026-08-28.

**Nothing here has been started.** Each PR below is specified to the point where it
can be picked up without re-deriving the research.

---

## The guardrail problem, and why it comes first

```
Go tests:  142 across 11 packages
CI runs:   go vet ./...  +  go test -race -cover ./...
JS tests:  none
CSS tests: none
JS lint:   none
```

Every UI defect in ROADMAP.md Track 1 is the kind CI could have caught and
didn't — a broken favicon path on all four pages, a CDN URL, 7.6 MB of
unreferenced PNGs, a `timeRange`/`timeframe` typo. And every UI change made so
far was verified by a human in a browser, which does not survive into CI.

B1 proved the point twice over on its first run: it found two defects the
six-agent audit missed (the favicon, and `.DS_Store` embedded into every
binary), and it disproved one the audit had reported (`js/core/main.js`, whose
three references are all inside HTML comments and therefore fetch nothing).

Tracks 2 and 3 are a **nine-step CSS/JS migration**. Doing that with no automated
verification is the largest risk in this plan. So **B1 lands first** and adds
guardrails that are pure Go, zero new dependencies, and run inside the existing CI
job.

---

## Milestones

| | Milestone | Closes | PRs |
| --- | --- | --- | --- |
| **A** | Release reachability | — | 1 |
| **B** | Guardrails + correctness | part of #32 | 6 |
| **C** | Design system | most of #32 | 9 |
| **D** | Accessibility | rest of #32 | 2 |
| **E** | Flame graphs | #43 | 2 |
| — | Deferred | #42, new IA issue | — |

Milestone letters are used rather than version numbers because Track 0 may
renumber the release line. If option B in ROADMAP.md is taken, A ships as
`v1.3.0`, B as `v1.4.0`, C as `v1.5.0`, D as `v1.6.0`.

---

## Milestone A — Release reachability

Blocks nothing technically, but until it lands **no user can install any of this**.

### A1 · Renumber the release line and document the release process

- **Branch** `chore/release-line-v1`
- **Files** `CHANGELOG.md`, `CONTRIBUTING.md`, `README.md`
- **Scope**
  - Renumber `[2.1.0]` → `[1.3.0]`, and annotate the `[2.0.0]` entry as never
    published to the module proxy, with the reason.
  - Add a "Releasing" section to `CONTRIBUTING.md`: tag, then **verify the proxy
    actually serves it** with the `curl` shown in ROADMAP.md Track 0. This is the
    step whose absence caused the problem.
  - README install snippet confirmed to reference a version that resolves.
- **Not in scope** Renaming the module to `/v2`. That is the alternative to this
  PR, not a follow-up.
- **Tests** None — docs only.
- **Acceptance** After tagging `v1.3.0`, `curl
  https://proxy.golang.org/github.com/iyashjayesh/monigo/@latest` returns
  `v1.3.0`, and `go get github.com/iyashjayesh/monigo@latest` in a scratch module
  pulls it.
- **Effort** Small. **Risk** Low, but it is a public versioning decision — needs a
  maintainer call before the PR, not during review.

---

## Milestone B — Guardrails and correctness

Six PRs. B1 first; B2–B6 are independent of each other afterwards and can land in
any order or in parallel.

### B1 · Asset guardrails in CI

- **Branch** `test/static-asset-guardrails`
- **Files** new `static_assets_test.go`, `.github/workflows/test.yml`,
  `static/index.html`, `static/function-metrics.html`
- **Scope** Four checks, all pure Go against the existing `staticFiles embed.FS`:
  1. **Every referenced asset exists.** Parse `src=` / `href=` out of each
     embedded `.html`, strip commented-out regions first, resolve relative paths,
     `fs.Stat` each one. Found the broken favicon on all four pages — that is the
     point. Stripping comments matters: without it, three commented-out
     `js/core/main.js` tags read as live references and produce a false positive.
  2. **No external URLs in embedded assets.** Reject `http://` and `https://` in
     `src`/`href`/CSS `url()` across `static/`, with an explicit allowlist that
     starts empty. Locks in the offline guarantee once B3 removes the CDN.
  3. **Payload budget.** Sum the embedded tree; fail over a constant. Set to the
     current size in this PR, tightened by B2. Stops 7.6 MB creeping back. Note
     the real embedded figure is **17.1 MB** summed from file sizes; `du -sk`
     reports 16.5 MB because it counts disk blocks.
  5. **No junk files.** `.DS_Store` and friends compiled into users' binaries.
  4. **JS parses.** A CI step running `node --check` over `static/js/*.js`
     excluding the vendored `echarts.min.js` and `backend-bundle.min.js`. Node is
     already on the GitHub runner; no `package.json`, no npm install.
- **Also** Fix the favicon path on all four pages and delete the embedded
  `.DS_Store` files, so checks 1 and 5 pass.
- **Tests** The PR *is* tests. Include a deliberately-failing fixture in review
  notes so a reviewer can confirm each check actually fires.
- **Acceptance** `go test ./... -race` green; temporarily reintroducing a CDN URL,
  a missing asset, or a syntax error each fails CI.
- **Effort** Medium. **Risk** Low. **Value** Highest per line in this plan.

### B2 · Remove unreferenced embedded assets

- **Branch** `fix/drop-unreferenced-static-assets`
- **Files** delete `static/assets/ss/d1–d10.png`, `static_assets_test.go` (budget)
- **Scope** Delete 7.6 MB of PNGs referenced by nothing. Tighten B1's payload
  budget to just above the new size. Check `.DS_Store` files under `static/` while
  in here — they are also embedded.
- **Tests** Budget assertion from B1 tightened.
- **Acceptance** `du -sk static` drops from 16896 KB to ~9100 KB; dashboard renders
  identically in all four pages.
- **Note for the PR body** This does not shrink existing clones — the blobs remain
  in git history — but the module zip is what `go get` downloads.
- **Effort** Small. **Risk** Low.

### B3 · Replace the icon layer with an inline SVG sprite

- **Branch** `fix/inline-icon-sprite`
- **Files** all four `static/*.html`, new `static/icons.html` partial or an inline
  block, `static/css/monigo-styles.css`, `static/js/common.js`
- **Scope**
  - Eight `<symbol>` elements: `check-circle`, `chevron-right`, `download`,
    `exclamation-triangle`, `github`, `refresh`, `moon`, `sun`. Sourced from a
    permissive-licence set (Lucide is ISC — copying eight paths is fine; **adding
    the library is not**), attribution in a CSS comment.
  - Remove the Font Awesome `cdnjs` `<link>` from all four pages.
  - Replace `fa-*` usages, and the dead `las la-bars` / `ri-menu-line` /
    `ri-menu-3-line` hamburger icons — **this is what restores mobile navigation**.
  - `common.js` theme toggle swaps `<use href>` rather than a font class.
  - Decide `html2canvas`: vendor it (~200 KB, against the payload budget) or drop
    the screenshot buttons. Recommendation: vendor it, because "Download Dashboard
    Image" is a genuinely useful thing to paste into an incident channel.
- **Tests** B1 checks 1 and 2 now pass with an empty allowlist.
- **Acceptance** No network requests to any external host on any page (verify with
  devtools offline mode); all eight icons render in both themes; sidebar opens
  below 1300px and the navbar below 992px.
- **Effort** Medium. **Risk** Medium — touches every page's markup. Screenshot
  before/after each page in the PR.

### B4 · Fix memory chart correctness

- **Branch** `fix/memory-chart-raw-values`
- **Files** `models/apiModels.go`, `core/core.go`, `api/api_test.go`,
  `static/js/index.js`
- **Scope**
  - Expose the existing raw byte counts (`HeapAllocByServiceRaw`,
    `HeapAllocBySystemRaw`, `TotalAllocByServiceRaw`, `TotalMemoryByOSRaw`) which
    currently carry `json:"-"`. Add **new** JSON field names alongside the
    formatted strings — additive, nothing existing breaks.
  - Memory Distribution pie and Heap Memory Usage bar read raw bytes and format
    only for display, instead of `parseFloat` on `"6.82 MB"`.
  - Same treatment for the CPU/load `parseFloat` calls at `index.js:553-557`
    if they suffer the same unit-stripping — confirm during implementation.
  - CPU Statistics pie: stop plotting `total_cores` as a slice alongside the two
    usage figures (it is the denominator, not a category).
- **Tests** Go test asserting the raw fields are present and non-zero in the
  `/metrics` response. A JS unit test is not possible without test
  infrastructure — verify in-browser and attach numbers to the PR.
- **Acceptance** A service using ~7 MB on a 16 GB host shows a visually
  proportionate slice, not ~30%. Compare against `runtime.ReadMemStats` output
  quoted in the PR body.
- **Effort** Medium. **Risk** Low-medium — API surface grows; keep it additive.

### B5 · Chart and range fixes

- **Branch** `fix/chart-range-and-resize`
- **Files** `static/js/reports.js`, `static/js/index.js`,
  `static/js/historycharts.js`, `static/js/goroutines.js`
- **Scope**
  - `reports.js:100` `timeRange` → `timeframe` (ReferenceError on the 7d range).
  - Pass `notMerge: true` to `setOption` so switching metric stops leaving stale
    series behind.
  - Add resize handling to the nine ECharts instances that lack it; one shared
    helper rather than nine listeners.
  - Default history window: make it longer than the sample interval, or derive it
    from the configured interval, so a chart's first render is not empty.
- **Tests** `node --check` from B1. Browser verification per fix in the PR body.
- **Acceptance** 7d range returns data; switching metric replaces series; charts
  reflow on window resize and rotation; first paint shows data.
- **Effort** Small. **Risk** Low.

### B6 · Interim contrast fix for the health number

- **Branch** `fix/health-band-contrast`
- **Files** `static/js/index.js`, `static/css/monigo-styles.css`
- **Scope** Only if Milestone C slips. Replace the five band colours from the
  vendored Grafana palette with values that clear 4.5:1 on white, and fix the
  `--%` gauge placeholder (currently 2.54:1). Use the semantic tokens from
  ROADMAP.md Track 2 so this is not thrown away by C.
- **Tests** Contrast ratios computed and quoted in the PR body for all five bands.
- **Acceptance** All five bands ≥ 4.5:1 in light theme; dark theme unchanged
  (already passing).
- **Effort** Small. **Risk** Low. **Skip if C1–C3 land within the same cycle.**

---

## Milestone C — Design system

Nine PRs, in order. **The dashboard renders correctly after every one**, and each
is independently revertable. C1–C5 are mechanical with zero-to-small visual diff;
**C6–C9 change what the dashboard looks like and does, and want separate review.**

Full token values, verified contrast ratios, and the chart palette are in
ROADMAP.md Track 2 and the design research behind it.

| PR | Branch | Scope | Visual diff | Effort |
| --- | --- | --- | --- | --- |
| **C1** | `feat/design-tokens` | Add `:root` + `body.dark-theme` token blocks at the top of `monigo-styles.css`. Change nothing else. | **None** | S |
| **C2** | `refactor/tokenize-cards-tables` | Replace hex literals with tokens for cards and tables. | **Small** (see note) | M |
| **C3** | `refactor/tokenize-chrome` | Same for sidebar, navbar, buttons, leak panel. | None | M |
| **C4** | `refactor/collapse-dark-overrides` | Delete each `body.dark-theme` override as its light counterpart becomes tokenized. ~400 of ~440 lines go. | None | M |
| **C5** | `refactor/borders-not-shadows` | Dark mode uses borders. Drop hover shadows and the `translateY(-2px)` card lift. | Small | S |
| **C6** | `feat/type-and-density` | Six-size type scale; `tabular-nums` on every metric cell; 34px table rows with sticky header; code blocks fixed-size with `overflow-x: auto`, never wrapped. | **Yes** | M |
| **C7** | `feat/charts-from-tokens` | ECharts theme reads tokens via `getComputedStyle`; toggle recolours without reload. | Small | M |
| **C8** | `feat/single-health-encoding` | Retire the four competing health encodings, keep one. Replace the 300s `location.reload()` with in-place fetches. | **Yes** | L |
| **C9** | `feat/loading-error-stale-states` | Accent bar after 300ms rather than a spinner over data; inline errors with the real message; one "as of" chip going amber past 2× and critical past 5× the poll interval; charts gap rather than flatline. | **Yes** | M |

### Notes that change how these are reviewed

- **C2 does not have zero visual diff, and cannot delete `!important` yet.** Both
  claims in the original plan were wrong, and testing found it. The tokens are a
  *corrected* palette, not a restatement of the current literals: 6 of the 8
  dominant values shift. The moves are small -- most under 20 on a 0-441 scale --
  and every affected text pair gains contrast (dark muted text 5.78:1 -> 8.38:1),
  but "None" was the wrong expectation to set for a reviewer. On `!important`:
  removing all 64 flags from the card and table rules changes nothing in light
  theme, but 7 elements in dark depend on one of them
  (`body.dark-theme .card { color }`), where the vendored `.bg-success` rule wins
  otherwise. Removing them safely needs per-declaration verification, and is
  better done in C4 once the vendored dark overrides are gone.
- **Measuring a CSS change needs a settle delay after any theme toggle.** An
  earlier C2 measurement reported 21 changed elements; the fingerprint had been
  captured mid-transition, and the "change" was the theme finishing. With a
  1s settle it reproduces as 0. Any before/after comparison in C3-C9 needs the
  same care, or it will invent regressions.
- **C1 is the safest PR in this plan and the highest leverage.** Zero visual
  diff — it only adds declarations. Merge it early even if C2 stalls.
- **C4 is where the `!important` count should collapse.** Quote before/after counts
  in the PR body: currently `203`. If it has not dropped substantially, the
  tokenization in C2/C3 was incomplete.
- **C8 deletes UI an existing user may be attached to** — the gauge, the health
  banner, the coloured tag. Say so explicitly in the PR, with a screenshot of what
  goes away, and consider a note in the changelog under Changed rather than Fixed.
- **C8 also changes refresh behaviour.** In-place fetching means the page no longer
  reloads; if anything depended on that reload (it shouldn't), it breaks. Test with
  the service killed mid-session — the current behaviour navigates to the browser
  error page, the new behaviour must show a stale-data state instead.
- **C6 and C9 are the visible payoff.** If the project wants a "the dashboard looks
  different now" moment for a release, it is these two.

---

## Milestone D — Accessibility

Depends on C1 for the `--focus` token. 24 audit findings sit here, 8 must-fix.

### D1 · Restore focus and keyboard operability

- **Branch** `fix/focus-and-keyboard`
- **Scope**
  - Restore focus rings. `static/css/core/backend.css:11987` sets
    `.btn:focus { outline: none; box-shadow: none }` and nothing in
    `monigo-styles.css` replaces it — tab through the dashboard today and nothing
    moves. Add `:focus-visible` styling driven by `--focus`.
  - Make MoniGo's own controls real controls: the theme toggle (an `<a>` with no
    `href`) becomes a `<button>` with `aria-pressed`; the refresh control (an
    `aria-hidden` `<i>` with a click listener) becomes a labelled button; the
    collapsible goroutine stack rows (click-only `div`s) become buttons with
    `aria-expanded`.
- **Acceptance** Every interactive element reachable and operable by keyboard with
  a visible focus indicator in both themes.
- **Effort** Medium. **Risk** Low.

### D2 · Names, labels and announcements

- **Branch** `fix/labels-and-live-regions`
- **Scope** Label the four unlabelled `<select>` controls. Give the goroutine leak
  verdict a live region so it is announced. Make the info "i" tooltips
  keyboard-reachable and stop clipping their content with `white-space: nowrap`.
  Fix the leak badge (white on `#f97316`, 2.80:1). Heading order and landmarks
  across all four pages.
- **Acceptance** A keyboard-and-screen-reader pass over each page, findings
  recorded in the PR.
- **Effort** Medium. **Risk** Low.

---

## Milestone E — Flame graphs (#43)

Independent of B–D beyond B1. Two PRs, because the parsing change is valuable on
its own.

### E1 · Parse profiles in-process

- **Branch** `feat/pprof-in-process`
- **Scope** Replace `exec.Command("go", "tool", "pprof", ...)` with
  `github.com/google/pprof` as a library. Removes the Go-SDK-on-host dependency
  that `ViewFunctionMetrics` currently warns about, and removes the argument-
  injection surface that #48 had to defend with input validation.
- **Tests** Existing `ViewFunctionMetrics` tests should pass unchanged; add tests
  over a committed fixture profile.
- **Effort** Medium. **Risk** Medium — new dependency; check its transitive weight
  against the payload/module-size concern before committing to it.

### E2 · Render the flame graph

- **Branch** `feat/flame-graph-view`
- **Scope** Emit folded stacks as JSON from E1's parsed profile; render in the
  Function Metrics page. **Do not CDN-load `d3-flame-graph`** — same offline
  constraint as the icons, and B1 check 2 will fail. Hand-roll the renderer: a
  flame graph is nested rectangles, a hover label, and click-to-zoom. Uses the
  Track 2 chart palette and tokens.
- **Effort** Large. **Risk** Medium.

---

## Deferred, with reasons

### Information architecture → file as a new issue

ROADMAP.md Track 4. The audit's position is that MoniGo is a catalogue of runtime
metrics rather than an instrument for answering a question: no time correlation
between signals, four unlinked range selectors, prime real estate spent on four
constants. C8 and C9 treat symptoms. The shape question needs a design
conversation before any code, and it should not sit inside #32 — **file it
separately so #32 can close honestly** once B–D land.

### External DB storage (#42) → blocked on a decision

ROADMAP.md Track 6. The code is easy; the distribution question is not. Adding
`pgx` + `mongo-driver` makes every consumer download drivers they will never use.
Three options in the roadmap, including declining and documenting the `Storage`
interface as the extension point. **Do not start coding.** Interacts with
Milestone A, since separate modules need the versioning settled first.

---

## Dependency graph

```
A1 ─────────────────────────────────────────── independent, do first

B1 ──┬── B2
     ├── B3 ──────────── restores mobile nav
     ├── B4
     ├── B5
     └── B6 (skip if C lands soon)

C1 ──┬── C2 ── C3 ── C4 ── C5 ── C6 ── C7 ── C8 ── C9   (strictly ordered)
     └────────────────────────────────── D1 ── D2       (D needs --focus from C1)

B1 ── E1 ── E2                                          (independent of C and D)
```

Parallelisable: A1, the B2–B6 set, and E1 can all proceed at once after B1. The C
chain is strictly sequential by design — each step assumes the previous one's
tokens exist.

## How to verify UI work until there is better tooling

B1 gives CI syntax, asset and payload checks, but nothing that catches "this looks
wrong". Until that exists, every UI PR should carry:

- before/after screenshots of each affected page, **in both themes**;
- computed contrast ratios for any colour pair introduced or changed, with the
  numbers in the PR body rather than an assertion that it is fine;
- a devtools-offline check for any PR touching assets;
- for C5–C9, a check at 1280px, 992px and 375px widths, since the responsive
  findings are unfixed until C6.

The example app under `example/` with `WithStorageType("memory")` and a short
`WithDataPointsSyncFrequency` is the fastest way to get a populated dashboard.
