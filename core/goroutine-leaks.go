package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/internal/alerting"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/models"
)

const (
	// defaultStaleThreshold is how long a goroutine must stay blocked before it
	// is treated as stale. The runtime reports block duration in whole minutes,
	// so anything below a minute is not expressible.
	defaultStaleThreshold = 24 * time.Hour

	// defaultSnapshotWindow is how many periodic snapshots are retained. Growth
	// is only reported once the window is full, so a group must rise in every
	// one of them to be flagged -- a deliberately strict rule, because a
	// worker pool scaling up briefly is not a leak.
	defaultSnapshotWindow = 5

	// maxSuspiciousGroups caps how many offending groups a report carries, so a
	// pathological process cannot produce an unbounded API response.
	maxSuspiciousGroups = 20
)

// goroutineHeaderPattern matches the header line the runtime writes for each
// goroutine, capturing the id, the wait reason, and the optional block duration:
//
//	goroutine 21 [chan receive]:
//	goroutine 21 [chan receive, 47 minutes]:
//	goroutine 1 [running]:
//
// The duration is only present once the runtime considers a goroutine blocked
// long enough to be worth reporting (see runtime/traceback.go).
var goroutineHeaderPattern = regexp.MustCompile(`^goroutine (\d+) \[([^,\]]+)(?:, (\d+) minutes)?\]:`)

var (
	leakMu sync.RWMutex
	// staleThreshold is the configured stale cutoff.
	staleThreshold = defaultStaleThreshold
	// snapshotWindow is the number of snapshots growth is computed across.
	snapshotWindow = defaultSnapshotWindow
	// snapshots holds per-signature counts, oldest first, at most
	// snapshotWindow entries. Only the periodic collector appends to it.
	snapshots []map[string]int
)

// SetStaleGoroutineThreshold configures how long a goroutine must remain
// blocked before it is reported as stale. A non-positive value restores the
// default of 24h.
func SetStaleGoroutineThreshold(d time.Duration) {
	leakMu.Lock()
	defer leakMu.Unlock()
	if d <= 0 {
		staleThreshold = defaultStaleThreshold
		return
	}
	staleThreshold = d
}

// GetStaleGoroutineThreshold returns the configured stale cutoff.
func GetStaleGoroutineThreshold() time.Duration {
	leakMu.RLock()
	defer leakMu.RUnlock()
	return staleThreshold
}

// parseGoroutineHeader extracts the id, wait reason, and block duration from a
// goroutine block's first line. ok is false when the line is not a header.
func parseGoroutineHeader(block string) (id int, state string, blockedMinutes int, ok bool) {
	header := block
	if idx := strings.IndexByte(block, '\n'); idx >= 0 {
		header = block[:idx]
	}

	m := goroutineHeaderPattern.FindStringSubmatch(strings.TrimSpace(header))
	if m == nil {
		return 0, "", 0, false
	}

	id, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, "", 0, false
	}
	if m[3] != "" {
		// A malformed duration is not worth discarding the goroutine over.
		blockedMinutes, _ = strconv.Atoi(m[3])
	}
	return id, m[2], blockedMinutes, true
}

// callStackOf returns a goroutine block with its header line removed, so that
// two goroutines sitting at the same place in the program compare equal
// regardless of their ids or how long they have been waiting.
func callStackOf(block string) string {
	if idx := strings.IndexByte(block, '\n'); idx >= 0 {
		return strings.TrimRight(block[idx+1:], "\n")
	}
	return ""
}

// signatureOf hashes a call stack into a short, stable identifier. The full
// stack can run to kilobytes, and it is carried separately on the group.
func signatureOf(callStack string) string {
	sum := sha256.Sum256([]byte(callStack))
	return hex.EncodeToString(sum[:6])
}

// groupGoroutines collapses raw goroutine blocks into one group per distinct
// call stack, carrying the longest observed block duration for each.
func groupGoroutines(blocks []string) map[string]*models.GoroutineGroup {
	groups := make(map[string]*models.GoroutineGroup)

	for _, block := range blocks {
		_, state, blockedMinutes, ok := parseGoroutineHeader(block)
		if !ok {
			continue
		}

		callStack := callStackOf(block)
		sig := signatureOf(callStack)

		g, seen := groups[sig]
		if !seen {
			g = &models.GoroutineGroup{
				Signature: sig,
				State:     state,
				CallStack: callStack,
			}
			groups[sig] = g
		}
		g.Count++
		if blockedMinutes > g.BlockedMinutes {
			g.BlockedMinutes = blockedMinutes
			// Report the state of the longest-blocked member, which is the one
			// a developer chasing a leak cares about.
			g.State = state
		}
	}

	return groups
}

// recordSnapshot appends per-signature counts to the retained window, evicting
// the oldest entry once the window is full.
//
// Only the periodic collector may call this. Growth is meaningless unless
// snapshots are evenly spaced, and dashboard polling is driven by whoever
// happens to have a browser tab open.
func recordSnapshot(groups map[string]*models.GoroutineGroup) {
	counts := make(map[string]int, len(groups))
	for sig, g := range groups {
		counts[sig] = g.Count
	}

	leakMu.Lock()
	defer leakMu.Unlock()
	snapshots = append(snapshots, counts)
	if len(snapshots) > snapshotWindow {
		snapshots = snapshots[len(snapshots)-snapshotWindow:]
	}
}

// growthFor reports the change in a signature's count across the retained
// window, and whether it rose at every step.
//
// monotonic is false unless the window is full, so a freshly started service
// cannot report a leak from two data points.
func growthFor(sig string) (growth int, monotonic bool) {
	leakMu.RLock()
	defer leakMu.RUnlock()

	if len(snapshots) < snapshotWindow {
		return 0, false
	}

	first, ok := snapshots[0][sig]
	if !ok {
		// Absent from the oldest snapshot: it may be genuinely new rather than
		// growing, and treating a zero baseline as growth flags every
		// short-lived stack.
		return 0, false
	}

	rising := true
	prev := first
	for _, snap := range snapshots[1:] {
		cur := snap[sig]
		if cur <= prev {
			rising = false
		}
		prev = cur
	}

	return prev - first, rising
}

// buildReport evaluates grouped goroutines against the stale threshold and the
// retained growth window.
func buildReport(total int, groups map[string]*models.GoroutineGroup) *models.GoroutineLeakReport {
	threshold := GetStaleGoroutineThreshold()
	thresholdMinutes := int(threshold.Minutes())

	leakMu.RLock()
	retained, required := len(snapshots), snapshotWindow
	leakMu.RUnlock()

	report := &models.GoroutineLeakReport{
		TotalGoroutines:       total,
		StaleThresholdMinutes: thresholdMinutes,
		SnapshotsRetained:     retained,
		SnapshotsRequired:     required,
	}

	var suspicious []models.GoroutineGroup
	for sig, g := range groups {
		entry := *g
		entry.Growth, entry.Growing = growthFor(sig)
		entry.Stale = thresholdMinutes > 0 && entry.BlockedMinutes >= thresholdMinutes

		if entry.Stale {
			report.StaleGoroutines += entry.Count
		}
		if entry.Growing {
			report.GrowingGroups++
		}
		if entry.Stale || entry.Growing {
			suspicious = append(suspicious, entry)
		}
	}

	// Worst first: longest-blocked, then fastest-growing, then largest. The
	// signature tie-breaker keeps ordering deterministic for equal groups.
	sort.Slice(suspicious, func(i, j int) bool {
		a, b := suspicious[i], suspicious[j]
		if a.BlockedMinutes != b.BlockedMinutes {
			return a.BlockedMinutes > b.BlockedMinutes
		}
		if a.Growth != b.Growth {
			return a.Growth > b.Growth
		}
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Signature < b.Signature
	})

	if len(suspicious) > maxSuspiciousGroups {
		suspicious = suspicious[:maxSuspiciousGroups]
	}
	report.SuspiciousGroups = suspicious
	report.LeakSuspected = report.StaleGoroutines > 0 || report.GrowingGroups > 0
	report.Message = describeReport(report, threshold)

	return report
}

// describeReport renders the human-readable summary carried on the report and
// used as the alert message.
func describeReport(r *models.GoroutineLeakReport, threshold time.Duration) string {
	if !r.LeakSuspected {
		if r.SnapshotsRetained < r.SnapshotsRequired {
			return fmt.Sprintf(
				"No goroutine leak detected across %d goroutines. Growth tracking warming up (%d/%d snapshots).",
				r.TotalGoroutines, r.SnapshotsRetained, r.SnapshotsRequired,
			)
		}
		return fmt.Sprintf("No goroutine leak detected across %d goroutines.", r.TotalGoroutines)
	}

	var parts []string
	if r.StaleGoroutines > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d goroutine(s) blocked for at least %s", r.StaleGoroutines, formatThreshold(threshold),
		))
	}
	if r.GrowingGroups > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d call stack(s) growing across the last %d snapshots", r.GrowingGroups, r.SnapshotsRequired,
		))
	}
	return fmt.Sprintf(
		"Possible goroutine leak: %s (total goroutines: %d).",
		strings.Join(parts, "; "), r.TotalGoroutines,
	)
}

// formatThreshold renders a duration the way an operator would say it. Whole
// hours read as hours; anything else stays in minutes rather than being rounded
// into a wrong-looking hour count.
func formatThreshold(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// leakReportFor evaluates already-captured goroutine blocks without recording a
// snapshot. Used by the on-demand API path, which is driven by dashboard
// polling and must not disturb the evenly-spaced growth history.
func leakReportFor(blocks []string) *models.GoroutineLeakReport {
	return buildReport(len(blocks), groupGoroutines(blocks))
}

// AnalyzeGoroutineLeaks captures every goroutine stack, records a snapshot for
// growth tracking, and returns the resulting verdict. A suspected leak raises
// an alert webhook if one is configured.
//
// This is the periodic entry point and is called from the metrics collection
// path, so detection does not depend on anyone having the dashboard open.
func AnalyzeGoroutineLeaks() *models.GoroutineLeakReport {
	blocks := SplitGoroutines(captureAllStacks())
	groups := groupGoroutines(blocks)

	recordSnapshot(groups)

	report := buildReport(len(blocks), groups)
	if report.LeakSuspected {
		logger.Log.Warn("possible goroutine leak detected",
			"stale_goroutines", report.StaleGoroutines,
			"growing_groups", report.GrowingGroups,
			"total_goroutines", report.TotalGoroutines,
		)
		alerting.TriggerAlert(common.GetServiceName(), float64(report.TotalGoroutines),
			"Goroutine Leak Alert: "+report.Message)
	}
	return report
}

// ResetLeakDetectionState clears the retained snapshot window and restores the
// default threshold. Exported for tests and for callers restarting collection.
func ResetLeakDetectionState() {
	leakMu.Lock()
	defer leakMu.Unlock()
	snapshots = nil
	staleThreshold = defaultStaleThreshold
	snapshotWindow = defaultSnapshotWindow
}
