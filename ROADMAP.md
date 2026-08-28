# MoniGo Development Roadmap

Living plan. Last updated 2026-08-28, after `v2.1.0` was cut in `CHANGELOG.md`.

Everything below is either a verified measurement or an explicitly-labelled
judgement call. Where a number appears, the command that produced it is given so
it can be re-checked rather than trusted.

---

## Where things stand

Merged and released as `2.1.0` in the changelog:

| PR | What |
| --- | --- |
| #48 | Hardening: inverted health score, unbounded in-memory growth, `WithStorageType` no-op, threshold data race, storage/goroutine/fd leaks, constant-time credential comparison, pprof argument injection |
| #49 | Opt-in health-breach webhook alerts (`WithAlertWebhook`) |
| #50 | Dashboard dark mode, skeleton loading states, grouped goroutine stacks |
| #51 | Goroutine leak detection (stale blocks + monotonically growing call stacks) |

Open issues: **#32** (dashboard redesign), **#42** (external DB storage), **#43**
(embedded flame graphs). #44 closed by #51; #45, #46, #47 closed.

---

## Track 0 — Release plumbing

**This gates everything else and should be settled first.** None of the work
above currently reaches anyone installing MoniGo.

### The problem

`v2.0.0` is tagged in git, but the Go module proxy refuses it:

```
$ curl -s https://proxy.golang.org/github.com/iyashjayesh/monigo/@latest
{"Version":"v1.2.0","Time":"2026-01-23T05:37:43Z", ...}

$ curl -s https://proxy.golang.org/github.com/iyashjayesh/monigo/@v/v2.0.0.info
not found: invalid version: module contains a go.mod file, so module path must
match major version ("github.com/iyashjayesh/monigo/v2")
```

`go.mod` declares `module github.com/iyashjayesh/monigo` with no `/v2` suffix, so
`v2.x` tags are invalid for module consumers. **Everyone running `go get` today
gets `v1.2.0` from January.** The OTel exporter, structured logging, graceful
shutdown, and every fix in `2.1.0` are unreachable. Tagging `v2.1.0` would land in
the same hole.

### The decision

| Option | Cost | Consequence |
| --- | --- | --- |
| **A. Rename module to `.../v2`** | Import-path change in every example, the README, and for anyone consuming v2 via a `replace` directive | Correct SemVer; `v2.x` tags become installable; the existing `v2.0.0` tag stays broken but is superseded |
| **B. Renumber this release `v1.3.0`** | The `[2.0.0]` changelog entry becomes historically odd | Installable immediately, zero import churn; abandons the 2.x line as never-published |

**Decision: B.** The release line is renumbered to `v1.3.0`, the `[2.0.0]`
changelog entry is annotated as never published, and `CONTRIBUTING.md` now carries
a Releasing section whose verification step is the one whose absence caused this.

A remains available later, if a genuinely breaking change ever justifies a major
bump. It was rejected for now because `/v2` would be permanent in every import
path for an API that is not v2-shaped, and per the proxy there were no v2
consumers to preserve.

**Still outstanding: the tag itself.** Renumbering the changelog does not publish
anything. Someone has to run `git tag -a v1.3.0` and confirm the proxy serves it,
per the checklist in `CONTRIBUTING.md`. Until that happens `go get` continues to
serve `v1.2.0`.

---

## Track 1 — Correctness defects

Six defects found during the UI audit. **These are bugs, not design**, and should
land before any redesign work so the redesign is not built on top of them. One PR,
or two if the memory fix needs API discussion.

### 1.1 Memory charts display mathematically wrong values

`static/js/index.js:688` and `:553-557` apply `parseFloat` to pre-formatted
strings:

```js
value: parseFloat(data.memory_statistics.memory_used_by_service)   // "6.82 MB" -> 6.82
value: parseFloat(data.memory_statistics.memory_used_by_system)    // "15.2 GB" -> 15.2
```

The unit is stripped, so 6.82 MB and 15.2 GB are plotted against each other in the
same pie. The Heap chart mixes KB and MB under an axis labelled `MB`.

A monitoring tool that misreports memory is worse than one that reports nothing,
so this is the highest-priority item on the list.

**Fix.** The raw byte counts already exist on `models.ServiceStats`
(`HeapAllocByServiceRaw`, `HeapAllocBySystemRaw`, `TotalAllocByServiceRaw`,
`TotalMemoryByOSRaw`) but carry `json:"-"`. Expose them — additively, under new
JSON names so nothing existing breaks — and have the charts read raw bytes,
formatting only for display. Add a Go test asserting the raw fields are present
and non-zero in the API response.

### 1.2 Mobile navigation is impossible, in every network condition

Both hamburger controls reference icon fonts that are never loaded:

```
static/index.html:32    <i class="las la-bars wrapper-menu"></i>       Line Awesome
static/index.html:126   <i class="ri-menu-line wrapper-menu"></i>      Remix Icon
static/index.html:154   <i class="ri-menu-3-line"></i>                 Remix Icon
```

`font-family: remixicon` appears in the vendored CSS, but there is **no
`@font-face` rule and no font file anywhere** under `static/` — no `.woff`,
`.woff2` or `.ttf`. These render as zero-size empty elements always, not just
offline. Below 1300px the sidebar cannot be opened; below 992px the navbar cannot
either.

**Fix.** Covered by 1.3 — replacing the icon layer restores navigation.

### 1.3 Icons: remove the CDN dependency

All four pages load Font Awesome 4.7.0 from `cdnjs.cloudflare.com`, and
`index.html` also CDN-loads `html2canvas`. On an airgapped host — which is where a
lot of production Go services live — every icon is a blank box.

Total glyphs actually used: **eight.** `check-circle`, `chevron-right`,
`download`, `exclamation-triangle`, `github`, `refresh`, `moon`, `sun`.

**Fix.** An inline SVG `<symbol>` sprite in a shared partial, roughly 2 KB
embedded. Replaces the CDN CSS, the never-loaded Line Awesome and Remix Icon
references, and the missing font files in one change. Vendor `html2canvas` or drop
the screenshot feature.

### 1.4 A 404 on every dashboard load: the favicon

All four pages carry `<link rel="shortcut icon" href="../assets/favicon.ico">`.
The pages are served from the root, so `../assets/` resolves *above* the served
tree and 404s on every load. The file does exist, at
`static/assets/favicon.ico` — the path just needs to be `./assets/favicon.ico`.

**Correction.** An earlier revision of this document claimed the 404 was
`js/core/main.js`, referenced by two pages and absent from disk. That was wrong:
all three references to it are inside HTML comments
(`<!-- <script src="./js/core/main.js" defer></script> -->`), so they fetch
nothing. The audit agent reported it and I confirmed only that the file was
absent and the string was present, without checking whether the reference was
live. The favicon above is the real instance of this defect class, and it was
found by the guardrail test in B1 rather than by the audit.

### 1.5 Reports "7d" throws a ReferenceError

`static/js/reports.js:100` reads `timeRange` where every sibling branch reads
`timeframe`. Selecting the 7-day range throws and blanks the table. One-word fix.

### 1.6 Drop 7.6 MB of unreferenced assets

```
$ du -sk static           16896 KB   (16.5 MB)
$ du -sk static/assets/ss  7772 KB   ( 7.6 MB)
$ grep -rn "assets/ss" static/ README.md      # no matches
```

`//go:embed static/*` (`monigo.go:32`) compiles all of it into every user's binary
and module download. `static/assets/ss/d1–d10.png` is referenced by no page, no
JS, no CSS, and not the README. Delete.

Note this does **not** shrink existing clones — the blobs stay in git history —
but the module zip is what every `go get` downloads.

### 1.7 Contrast: the health number, and one bug from #50

Measured live via `getComputedStyle` on a running dashboard. The gauge percentage
is coloured by band from the vendored Grafana palette:

| Band | Token | Value | On white | |
| --- | --- | --- | --- | --- |
| ≥80 Excellent | `--green` | `#37872d` | 4.50 | pass |
| ≥60 Good | `--lightgreen` | `#56a64b` | 3.02 | fail |
| ≥50 Moderate | `--yellow` | `#eab839` | **1.84** | fail |
| ≥40 Degraded | `--orange` | `#ef843c` | 2.61 | fail |
| <40 Critical | `--red` | `#f2495c` | 3.57 | fail |

Four of five states fail AA in light theme, and the two that matter during an
incident are the worst — the dashboard becomes least legible exactly when
something is wrong. `--green` also resolves to `#37872d` on `body` but `#78c091`
on `:root`: two competing definitions inherited from the template.

Separately, the gauge placeholder added in #50 (`--%` at `#9ca3af` on white) is
**2.54:1** and fails. Both are fixed by the Track 2 token layer, but if Track 2
slips, fix these two independently.

---

## Track 2 — Design system (issue #32)

Full research, reference comparison and the verified token set sit behind this
section; the PR-by-PR execution order is in
[DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md) (milestone C). Summary and decisions
below.

### Direction

**Primary reference: ClickHouse**, of twelve `DESIGN.md` systems evaluated. The
only candidate shipping a real five-plane dark elevation ramp built from pure
neutrals, so it inverts to light mode mechanically rather than requiring every hue
to be re-picked. Its depth-from-hairlines-not-shadows doctrine, fixed-code-size
rule, and one-big-number stat pattern are directly what MoniGo needs. Its yellow
`#faff69` is rejected: yellow-as-brand beside amber-as-warning is the worst
legibility failure available to an observability UI.

Three governing doctrines:

1. **One fact, one encoding.** Health is currently asserted four ways — a gauge, a
   percentage, a coloured tag, and a top banner — which can disagree with each
   other. Keep one.
2. **Colour is never the only channel.** Every status carries a glyph, a word and a
   colour; every chart series a direct label or dash pattern.
3. **The accent is identity, never state.** `#ff5c35` never enters the chart
   palette and never fills a data surface.

### Issue #32's open decisions, resolved

| Question | Answer | Why |
| --- | --- | --- |
| Bootstrap 5 vs Tailwind vs hybrid | **None of the three.** A token layer over existing Bootstrap 4 | Tailwind needs a build step; the UI is served from `embed.FS` with no npm, bundler or `package.json`. Bootstrap 5 is migration cost that fixes none of the real defects |
| Lucide vs Heroicons vs FA6 | **None of the three.** Eight inline SVG symbols | Eight glyphs are in use. A sprite is ~2 KB; the smallest library is two orders of magnitude more, permanently, for every module consumer |
| Dark mode priority | Shipped in #50; now the strongest argument for tokens | It cost 51 selectors and ~440 lines of overrides because there was no token layer to invert |
| Mobile-first commitment | **No — desktop-dense first.** But mobile-*working* is non-negotiable | The user is an operator on a large monitor, deliberately. A centred 1280px column wastes half of it. Mobile-first optimises for the wrong primary case; today's mobile is nonetheless broken (1.2) |

### The measured case for tokens

```
$ grep -oE '#[0-9a-fA-F]{3,8}' static/css/monigo-styles.css | wc -l    105
$ grep -oE '#[0-9a-fA-F]{3,8}' static/css/monigo-styles.css | sort -u | wc -l    22
$ grep -c '!important' static/css/monigo-styles.css                    203
$ grep -c 'body\.dark-theme' static/css/monigo-styles.css               51
```

Eight values account for 82 of the 105 literals (`#1f2937`×21, `#ff5c35`×17,
`#f3f4f6`×12, `#374151`×9, `#e5e7eb`×7, `#9ca3af`×7, `#ffffff`×6, `#111827`×6),
and every one is a 1:1 token substitution. **203 `!important` declarations is the
real disease** — they exist because there was no cascade layer to win with. Around
400 of the ~440 dark-override lines collapse (~91%), leaving a small block of
genuine structural exceptions.

Also worth knowing: `monigo-styles.css` already reaches for `'Inter'` (line 338)
and `'Fira Code'` (line 318). Neither ships, so it silently falls back today while
encoding the wrong font metrics.

### Migration order

The dashboard renders correctly after every step, and each is independently
revertable. Steps 2–5 are mechanical; **6–9 change appearance and behaviour and
deserve separate review.**

1. Track 1 defects first (above) — bugs before design.
2. Add the `:root` + `body.dark-theme` token blocks. Change nothing else. Zero
   visual diff, reviewable and revertable on its own.
3. Tokenize light theme one section at a time: cards → tables → sidebar/navbar →
   buttons → leak panel. Delete each `!important` where the token now wins unaided.
4. Delete each dark override as its light counterpart lands. This is where the
   ~400 lines go. Same order, so mistakes stay scoped to one component.
5. Shadows out, borders in. Drop hover shadows and the `translateY(-2px)` card
   lift, so a 20-tile grid stops bouncing under the cursor.
6. Type and density: six sizes, `tabular-nums` on every metric cell so live
   columns stop jittering on each poll, 34px table rows with a sticky header, code
   blocks at fixed size with `overflow-x: auto` and never wrapped.
7. Charts read from tokens via `getComputedStyle`, so the theme toggle recolours
   without a reload. Fix the nine ECharts instances that never resize, and pass
   `notMerge` so switching metric stops leaving stale series behind.
8. Retire redundant encodings. Keep one health representation: value, sparkline,
   threshold-coloured text. Replace the 300s `location.reload()` with in-place
   fetches so investigation state survives. **This step deletes UI an existing
   user may be attached to — flag it in the PR.**
9. Loading, error and stale states, built from the tokens that now exist: a 2px
   accent bar on the card chrome after a 300ms delay rather than a spinner over
   the data; errors inline with the real message; one "as of" chip that goes amber
   past 2× the poll interval and critical past 5×. Charts gap at the last real
   sample rather than flatlining. A stale value never keeps a green tag.

### Constraints that must not be violated

- **No webfont, including "just Inter" and "just the mono".** Every byte ships in
  every consumer's binary, and CDN fonts do not resolve on airgapped hosts.
  Corollary: cap negative letter-spacing at −0.6px and only above 24px, and cap
  weight at 600 — heavier weights synthesise unpredictably on Linux browsers.
- **No icon library after removing Font Awesome.** Hand-write the ninth glyph.
- **No marketing rhythm.** Spacing scale stops at 32px, type scale at 30px, no
  centred max-width container.
- **No build pipeline for the tokens.** If an improvement needs a `package.json`,
  it is not an improvement here.

---

## Track 3 — Accessibility pass (issue #32, remainder)

24 findings sit in this dimension, 8 of them must-fix. Distinct enough from the
token work to be its own PR, and it depends on Track 2 for the `--focus` token.

- Every interaction MoniGo itself added is mouse-only: the theme toggle (an `<a>`
  with no `href`), the manual refresh control (an `aria-hidden` `<i>` with a click
  listener), the collapsible goroutine stack rows (click-only `div`s), the info
  "i" tooltips (hover-only, keyboard-unreachable, content clipped by
  `white-space: nowrap`).
- `.btn:focus { outline: none; box-shadow: none }` in the vendored
  `static/css/core/backend.css` strips the focus ring from every button, and
  nothing in `monigo-styles.css` restores it. Tab through the dashboard today and
  nothing moves visually.
- Four functional `<select>` controls have no label.
- The goroutine leak verdict is injected with no live region, so it is never
  announced.
- Leak badge (white on `#f97316`) is 2.80:1 — the count an operator most needs to
  read.

---

## Track 4 — Information architecture

The open question, and the one with no obvious answer. Recorded rather than
scheduled.

The audit's position: MoniGo is currently organised as *a catalogue of every
metric the Go runtime exposes*, not as an instrument for answering a question. All
three incident scenarios end with the operator scrolling a 17-widget page whose
four history charts each carry their own unlinked time dropdown. Concretely:

- No time correlation between signals — each chart is an isolated widget with its
  own range selector, so "did the goroutine count spike when memory did" cannot be
  answered.
- Default history window (5m) is shorter than the default sample interval (5m), so
  every chart's first render is empty.
- The top-right eight columns of prime real estate hold four constants (service
  name, Go version, start time, PID) that can never be the reason anyone opened
  the page.
- Failure states: every fetch site swallows errors to `console.error`, nothing
  says how old the data is, and the 5-minute reload wipes state exactly when the
  service dies.

Worth its own issue and a design conversation before any code. Steps 8 and 9 of
Track 2 address the symptoms; this is the shape question.

---

## Track 5 — Embedded flame graphs (issue #43)

Self-contained, high visual payoff, no architectural risk. Good candidate once
Tracks 1–2 land.

Currently `ViewFunctionMetrics` shells out to `go tool pprof`, which means profile
views already fail when the Go SDK is absent from the host — a limitation the code
acknowledges in its own warning message.

Approach: parse profiles in-process with `github.com/google/pprof` rather than
shelling out (this also removes the Go-SDK dependency), emit folded stacks as
JSON, and render with a vendored flame-graph renderer. **Do not CDN-load
`d3-flame-graph`** — same offline constraint as the icons. Either vendor it or
hand-roll the renderer; a flame graph is rectangles and a hover label, and the
hand-rolled version avoids adding d3 to the embedded payload.

---

## Track 6 — External database storage (issue #42)

**Blocked on a design decision, not on effort.** Do not start coding this.

The `Storage` interface is already decoupled and tiny:

```go
type Storage interface {
    InsertRows(rows []Row) error
    Select(metric string, labels []Label, start, end int64) ([]DataPoint, error)
    Close() error
}
```

So implementing Postgres and MongoDB is straightforward plumbing. The problem is
distribution: MoniGo is a library embedded in other people's binaries, `go.mod`
already carries 44 requires including all of Fiber, and adding `pgx` +
`mongo-driver` makes every user pay for drivers they will never use.

Options to settle first:

- **Separate modules** — `monigo-storage-postgres`, `monigo-storage-mongo`, each
  with its own `go.mod`. Cleanest; users opt in by importing. Interacts with
  Track 0, so settle module versioning first.
- **Build tags** — keeps one module, but the drivers still appear in `go.mod` and
  still get downloaded.
- **Decline** — document the `Storage` interface as the extension point and let
  users implement their own. Zero cost, and arguably the right answer for a
  library whose selling point is that it is a drop-in.

Secondary design issue: `Select(metric, labels, start, end)` maps cleanly onto
neither SQL nor documents, and multi-instance deployments writing to one database
raise a labelling question (`host` is currently a single label) that the interface
does not currently express.

---

## Sequencing

PR-level breakdown, branch names, per-PR tests and acceptance criteria are in
[DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md). This section is the track-level view.

```
Track 0  release plumbing        ── settle first, one tag + a doc note
   │
Track 1  correctness defects     ── one PR, bugs before design
   │
Track 2  tokens, steps 2-5       ── mechanical, zero-to-small visual diff
   │        │
   │        └── steps 6-9        ── visible change, separate review each
   │
Track 3  accessibility           ── after Track 2 (needs --focus token)
   │
Track 5  flame graphs (#43)      ── independent, any time after Track 1
   │
Track 4  IA rework               ── needs a design conversation first
Track 6  external storage (#42)  ── blocked on the module decision
```

Tracks 1, 2 and 3 close the bulk of #32. Track 4 is what remains of it afterwards
and should probably become its own issue so #32 can close honestly.

## Notes on method

The audit behind Tracks 1–4 ran 12 agents across 6 dimensions, each dimension
paired with an agent instructed to refute rather than confirm: 107 raw findings,
104 retained. A second workflow of 6 agents evaluated 12 design references and
researched Grafana, Datadog, Sentry and pprof conventions.

Two agent findings were **wrong** and were withdrawn after local checking, both
originally ranked must-fix:

- *"The gauge arc never renders because `var()` is used in an SVG presentation
  attribute."* Probed live: `arc_computed_stroke` returns `rgb(55, 135, 45)`.
  Chrome resolves it. The arc renders and tracks the band.
- *"`.gauge` is already neutralised with `display:none`."* It is
  `display: block !important`; only `.gauge::after` is hidden.

All 22 contrast ratios and 16 chart-series values in the proposed token set were
independently recomputed and matched exactly. Treat any unverified agent claim in
future work the same way.
