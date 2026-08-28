# Adopting the new dashboard design

Source: `MoniGo Dashboard (1).html` (Claude Design canvas export, 598 KB
bundled), reviewed and patched 2026-08-28. The corrected export is
`MoniGo Dashboard (fixed).html`.

Companion to [ROADMAP.md](ROADMAP.md) and
[DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md). This supersedes milestone C in the
latter, which was written before the design existed.

Every value here was extracted from the file and every contrast ratio computed.

---

## 1. Three fixes applied

The revision fixed the blocker from the first export: the light theme now
carries its own semantic values, and all five pass AA (5.33–6.23). Three
defects remained. All are patched and verified in the browser in both themes.

| # | Defect | Was | Now |
| --- | --- | --- | --- |
| 1 | `--onacc: var(--onacc)` in `:root` — a self-referential custom property, invalid at computed-value time, so 8 dark accent buttons inherited `--txt` | 2.02:1 | `#04121A` → **7.84:1** |
| 2 | Dark `--dim2`, the most-used text tone at **90 usages** — nav headers, `pid 94917 · go1.21`, `412 MB / 1.0 GB retained` | 2.99:1 | `#778292` → **4.75:1** |
| 3 | `--line2` added for control boundaries, but used 0 times and below the 3:1 non-text floor in both themes | 1.68 / 1.88 | `#526683` / `#8191AB` → **3.16 / 3.20** |

Post-patch audit: **20 checks, 0 failures** across both themes — text on panel,
all five semantic tones, control borders, on-accent button labels. Both themes
confirmed rendering.

`--dim2` was tuned along its own hue rather than replaced, and the tonal
ordering still holds: `--txt` 15.45 > `--dim` 5.88 > `--dim2` 4.75.

**Take these three values back to the canvas** so the next export carries them.
The patched file is a stopgap, not the source of truth.

### A note on how the patch was made

The first attempt corrupted the file. `json.dumps` wrote a literal `</script>`
inside the embedded payload, which terminates the enclosing script tag early.
The original bundle escapes it as `</script>`; the patch now does the same,
and the output is verified to decode, retain all 25 assets, and render.

### A correction to my first review

I reported dark `--dim2` as 7.18:1 PASS. That was wrong — I had assigned the
light value (`#98A2B3`) to the dark theme. The real value was `#59626F` at
2.99:1, a failure on the most-used text tone in the design. My first review
therefore missed its largest accessibility defect.

### What the revision improved unprompted

Hardcoded `rgba()` tints became `color-mix(in srgb, var(--x) N%, transparent)`,
so a tint derives from its token and adapts per theme rather than being frozen
to the dark values. Better than what the review asked for.

---

## 2. The design's system, as built

```
                     DARK (:root)        LIGHT (body[data-theme="light"])
--bg                 #080B10             #F3F5F9
--pan                #0F141B             #FFFFFF
--pan2               #151B24             #F7F9FC
--in                 #0B0F15             #F7F9FC
--line   divider     #1D2531             #E3E8EF
--line2  control     #526683  (patched)  #8191AB  (patched)
--txt                #E7EBF1             #0E141B
--dim                #8892A2             #5B6674
--dim2               #778292  (patched)  #6B7686
--acc                #00B4E6             #00758F
--onacc              #04121A  (patched)  #FFFFFF
--ok                 #5FD39B             #0F7A4F
--warn               #FFA24B             #A15600
--bad                #FF6B7A             #C0273C
--vio                #8B7BF7             #5B49D6
```

Type: **IBM Plex Sans** + **IBM Plex Mono**, 21 bundled woff2 files, 230 KB.
Sizes 10–16px across nine steps, weights 400/500/600/700, no display tier.
Radii: 12 distinct values (2–14px).

Shell: 200px sidebar plus content. Navigation regrouped as
`MONITOR` (Overview · Functions · Goroutines · Reports),
`PIPELINE` (Exporters · Settings), `STATES`, `STORAGE`.

---

## 3. What is buildable now vs what needs Go work

This is the split that matters for planning. Several screens show data MoniGo
does not collect — those are features, not restyles, and should not be folded
into a CSS milestone.

### Buildable against today's API

| Design element | Backing |
| --- | --- |
| Whole shell, sidebar, nav grouping | frontend only |
| Overview metric tiles (PID, UPTIME, CORES, SERVICE CPU LOAD, HEAP IN USE, GOROUTINES, GC PAUSE) | all present on `/metrics` |
| `SYSTEM HEALTH` threshold bars (`198.33% / 90.00%`) | `core.GetThresholds()` + health scores, both shipped |
| Heap composition | `raw_mem_stats_records`, corrected in #55 |
| Runtime load multi-series chart | `load_statistics` |
| `NET SENT` / `NET RECV` / `STACK` / `SWAP FREE` | present |
| Goroutines: `LIVE GOROUTINES`, `BY STATE`, `LONGEST BLOCKED` | `#51` collects all of it |
| Leak table: `SIGNATURE / TOP FRAME / STATE / COUNT / BLOCKED / GROWTH` | exactly `models.GoroutineGroup` |
| `BY STATE` / `LONGEST BLOCKED` | needed a small Go addition after all: the report carried only *suspicious* groups, and a breakdown by state cannot be computed from the offenders alone. `GoroutineLeakReport.Groups` now carries every distinct stack, capped at 60 with `GroupsTotal` uncapped. |
| `2 stale groups, 1 growing across 6 of 6 retained snapshots` | `GoroutineLeakReport` verbatim |
| Disconnected / stale states | frontend only |
| `STORAGE` footer (`412 MB / 1.0 GB retained`, `disk · 7d`) | retention config + `common.GetDirSize` |
| Exporters page — Prometheus `SCRAPED`, OTel `RETRYING` + error text | partial: both exporters exist, but no status surface |

### Needs Go work first

| Design element | Missing |
| --- | --- |
| `P50` / `P95` per traced function | `FunctionMetrics` records totals, not percentiles. Needs a histogram or reservoir per function |
| Per-function sparkline / `+2.14 MB` delta column | no per-function time series |
| `Flame` / `Top` / `Graph` profile tabs | issue #43; profiles are currently shelled out to `go tool pprof` |
| `Generate PDF` on Reports | no renderer; the offline constraint rules out a hosted service |
| `RESOLUTION` control | storage has no downsampling |
| `Settings` page | configuration is builder-time only, not runtime-editable |
| `First run & empty` state | needs a "no data yet" signal the API does not expose |
| `Dashboard login` board | auth middleware exists; there is no login *page*, only 401s |

---

## 4. Adoption plan

Milestones A and B are done and unaffected. C1 and C2 are merged and survive —
the token layer's *structure* was right, and only the values change. What
follows replaces C3–C9.

### Status

| | Milestone | State |
| --- | --- | --- |
| D1 | Token values retargeted | done |
| D2 | Type and density | done |
| D3 | Shell and Overview | **done** — `feat/design-overview` |
| D4 | Goroutines page | **done** |
| D5 | States: disconnected, stale, empty | **done** — folded into each page's renderer |
| D6 | Reports and Exporters | **Reports done**; Exporters still needs the Go status surface |

Verified across all four pages, both themes: **566 text elements, zero
contrast failures.**

---

### Token corrections found while building — take these back to the canvas

Three values in §2 are wrong in the export and are corrected in the shipped
CSS. These are on top of the three from §1.

| Token | Export | Shipped | Why |
| --- | --- | --- | --- |
| `--ink-subtle` light | `#6B7686` | `#657080` | 4.60:1 on the white panel but **4.22 on `--canvas`** and 4.37 on `--surface-2`/`--inset`. It had been patched at three separate call sites before the pattern was clear. |
| `--ink-subtle` dark | `#778292` | `#798494` | 4.75:1 on `--pan` but **4.44 on `--surface-2`**, which table headers sit on. |

The lesson generalises: **a tone tuned against one surface is not safe on the
others.** The design has four grounds (`--surface-1`, `--canvas`, `--surface-2`,
`--inset`); a text token has to clear 4.5:1 on the darkest of them in light and
the lightest in dark, not just on the panel it was sampled against. Both values
above were tuned along their own hue, so the tonal ordering still holds:
`--ink` > `--ink-muted` > `--ink-subtle`.

A guardrail now covers the adjacent failure mode: `TestStylesheetsDefineEveryTokenTheyUse` fails the build on a `var(--x)` this stylesheet never defines.
An undefined custom property does not error — the browser drops the declaration
and the element silently keeps its inherited value — which is how the footer
briefly shipped using the design file's token names instead of the CSS's.

---

### D1 · Retarget the token values to the design

- **Branch** `feat/design-tokens-v2`
- Replace the values in the `:root` / `body.dark-theme` blocks with the design's,
  keeping the existing token *names* so C2's rules keep working. Add `--onacc`,
  `--line2`, `--pan2`, `--in`; map `--surface-1/2/3` onto `--pan`/`--pan2`/`--in`.
- Switch the theme mechanism from `body.dark-theme` to `body[data-theme]` to
  match the design, keeping the `localStorage` key.
- **Prerequisite:** finish the open bug on `refactor/collapse-dark-theme` first
  (see §5). Retargeting values while 44 dark overrides still exist means editing
  44 rules instead of one block.
- **Verify** by recomputing all 20 contrast checks against the shipped CSS, not
  just the design file.
- Effort S once the collapse lands. Visual diff: **large and intended.**

### D2 · Type and density

- IBM Plex, subject to the decision in §6. Nine sizes reduced to a documented
  scale; 12 radii reduced to three.
- `font-variant-numeric: tabular-nums` on every metric cell so live columns stop
  jittering on each poll.
- Effort M. Visual diff: large.

### D3 · Shell and Overview

- 200px sidebar, regrouped nav, the tile grid, sparkline tiles, `SYSTEM HEALTH`
  threshold bars, storage footer.
- This is the first PR where the dashboard looks like the design.
- Effort L.

### D4 · Goroutines page

- Leak table, `BY STATE`, `LONGEST BLOCKED`. **No API work needed** — this is
  the highest ratio of visible progress to effort in the whole plan, because
  `#51` already produces every column.
- Effort M.

### D5 · States: disconnected, stale, empty

- The disconnected banner naming the endpoint and the last-good timestamp;
  the "as of" staleness escalation; charts gapping rather than flatlining.
- Replaces the 300s `location.reload()` with in-place fetching, so investigation
  state survives — this was C8's most valuable half.
- Effort M.

### D6 · Reports and Exporters

- Reports to the new layout, minus `RESOLUTION` and `Generate PDF`.
- Exporters page showing real Prometheus and OTel status. Needs a small Go
  addition to expose exporter state.
- Effort M, plus a little Go.

### Then, as separate issues

Function percentiles; flame graph (#43); settings surface; login page. Each is a
feature with its own design questions, and each should get an issue rather than
riding along in a UI PR.

---

## 5. The sidebar-in-dark-mode bug — resolved

`.iq-sidebar` rendered white in dark mode even though the token resolved
correctly on the element and my override carried `!important`.

**Resolved:** the vendored template sets the shorthand
(`.iq-sidebar { background: #ffffff }`), and a `background-color` longhand with
`!important` does not beat a later shorthand in that cascade position. Matching
the shorthand form fixed it, and all four pages now render correctly in dark
mode — verified in the browser, not inferred.

The exact precedence mechanism was never established, only the fix. That is
worth knowing if it recurs on another vendored selector: **match the property
form the vendored rule uses**, rather than assuming `!important` on a longhand
is enough.

Also fixed on that branch: dark component rules 51 → 2, and the light theme's
skeleton text 2.54:1 → 5.43:1, filled buttons 3.07:1 → 5.78:1, active sidebar
link 3.07:1 → 5.18:1.

---

## 6. The one decision still open

**Ship IBM Plex, or keep system fonts?**

21 bundled woff2 files, 230 KB — **3.0%** of the payload as it now stands after
the reduction from 17.1 MB to 7.9 MB. They are bundled, not CDN-fetched, so they
work on an airgapped host; my earlier blanket objection to webfonts applied to
CDN delivery and was too broad.

- **Ship it, subset.** Renders as designed. Subsetting to Latin and the four
  weights actually used would cut the 230 KB materially.
- **Keep system fonts.** 0 KB. Nothing breaks, but metrics differ, so density
  and any tracking need re-checking and the drawn look will not be exact.
- **Plex Mono only.** Roughly half the bytes. Stack traces and pprof output are
  the primary content of two of the four pages, which is where a mono face earns
  its weight; UI text stays on the system stack.

Recommendation: **Plex Mono only**, as the best ratio of fidelity to bytes — with
full Plex a reasonable call if the design's exact look matters more than 230 KB.

**What ships today:** `--font-sans` and `--font-mono` name IBM Plex first, but
**no font file is bundled and there is no `@font-face` rule**, so every page
falls through to the system stack. That is a deliberate hold, not an oversight —
the decision above is still open — but it means the drawn result is *not* pixel-
identical to the design, and the density was checked against the fallback
metrics rather than Plex's. Bundling the face later will shift spacing slightly
and the density pass should be repeated then.

The accent question from the first review is settled by the design: cyan
`#00B4E6` dark / `#00758F` light. Worth noting it also removes a real problem —
orange-as-brand adjacent to amber-as-warning was the worst misread available to
an observability UI, and cyan has no such collision.
