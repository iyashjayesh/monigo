package core

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

// block renders a goroutine block the way runtime.Stack writes it.
func block(id int, state string, minutes int, fn string) string {
	header := fmt.Sprintf("goroutine %d [%s]:", id, state)
	if minutes > 0 {
		header = fmt.Sprintf("goroutine %d [%s, %d minutes]:", id, state, minutes)
	}
	return header + "\n" +
		fmt.Sprintf("main.%s(...)\n", fn) +
		"\t/app/main.go:42 +0x1c\n" +
		"created by main.main\n" +
		"\t/app/main.go:17 +0x88\n"
}

func TestParseGoroutineHeader(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantID      int
		wantState   string
		wantMinutes int
		wantOK      bool
	}{
		{"running, no duration", "goroutine 1 [running]:\nmain.main()", 1, "running", 0, true},
		{"chan receive with duration", "goroutine 21 [chan receive, 47 minutes]:\nx()", 21, "chan receive", 47, true},
		{"select with duration", "goroutine 9 [select, 1440 minutes]:\nx()", 9, "select", 1440, true},
		{"IO wait", "goroutine 5 [IO wait]:\nx()", 5, "IO wait", 0, true},
		{"semacquire", "goroutine 7 [semacquire, 3 minutes]:\nx()", 7, "semacquire", 3, true},
		{"sync.WaitGroup.Wait", "goroutine 8 [sync.WaitGroup.Wait, 12 minutes]:\nx()", 8, "sync.WaitGroup.Wait", 12, true},
		{"large id and duration", "goroutine 1234567 [chan send, 99999 minutes]:\nx()", 1234567, "chan send", 99999, true},
		{"header only, no body", "goroutine 3 [sleep]:", 3, "sleep", 0, true},
		{"not a header", "main.main()\n\t/app/main.go:9", 0, "", 0, false},
		{"empty", "", 0, "", 0, false},
		{"missing brackets", "goroutine 4 running:", 0, "", 0, false},
		{"non-numeric id", "goroutine abc [running]:", 0, "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, state, minutes, ok := parseGoroutineHeader(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if id != tt.wantID {
				t.Errorf("id = %d, want %d", id, tt.wantID)
			}
			if state != tt.wantState {
				t.Errorf("state = %q, want %q", state, tt.wantState)
			}
			if minutes != tt.wantMinutes {
				t.Errorf("minutes = %d, want %d", minutes, tt.wantMinutes)
			}
		})
	}
}

// The parser must handle real runtime output, not just the fixtures above.
func TestParseGoroutineHeaderAgainstRealRuntimeOutput(t *testing.T) {
	blocks := SplitGoroutines(captureAllStacks())
	if len(blocks) == 0 {
		t.Fatal("expected at least one goroutine block")
	}

	parsed := 0
	for _, b := range blocks {
		if _, _, _, ok := parseGoroutineHeader(b); ok {
			parsed++
		}
	}
	if parsed != len(blocks) {
		t.Errorf("parsed %d of %d real goroutine blocks; the header format may have changed", parsed, len(blocks))
	}
}

// Goroutines at the same place in the program group together regardless of id
// or how long each has been waiting.
func TestGroupGoroutinesCollapsesIdenticalStacks(t *testing.T) {
	blocks := []string{
		block(1, "chan receive", 0, "worker"),
		block(2, "chan receive", 5, "worker"),
		block(3, "chan receive", 90, "worker"),
		block(4, "select", 0, "dispatcher"),
	}

	groups := groupGoroutines(blocks)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	var worker, dispatcher *models.GoroutineGroup
	for _, g := range groups {
		if strings.Contains(g.CallStack, "worker") {
			worker = g
		} else {
			dispatcher = g
		}
	}

	if worker == nil || dispatcher == nil {
		t.Fatal("expected one worker group and one dispatcher group")
	}
	if worker.Count != 3 {
		t.Errorf("worker count = %d, want 3", worker.Count)
	}
	// The longest-blocked member's duration wins; that is the one worth chasing.
	if worker.BlockedMinutes != 90 {
		t.Errorf("worker blocked minutes = %d, want 90", worker.BlockedMinutes)
	}
	// ...and its state, since that is the goroutine a developer will chase.
	if worker.State != "chan receive" {
		t.Errorf("worker state = %q, want %q", worker.State, "chan receive")
	}
	if dispatcher.Count != 1 {
		t.Errorf("dispatcher count = %d, want 1", dispatcher.Count)
	}
}

// Blocks with no parsable header are skipped rather than counted.
func TestGroupGoroutinesSkipsUnparsableBlocks(t *testing.T) {
	groups := groupGoroutines([]string{
		block(1, "chan receive", 0, "worker"),
		"garbage without a header\n\tnot a stack",
		"",
	})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
}

func TestStaleThresholdConfiguration(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)

	if got := GetStaleGoroutineThreshold(); got != defaultStaleThreshold {
		t.Errorf("default threshold = %v, want %v", got, defaultStaleThreshold)
	}

	SetStaleGoroutineThreshold(90 * time.Minute)
	if got := GetStaleGoroutineThreshold(); got != 90*time.Minute {
		t.Errorf("threshold = %v, want 90m", got)
	}

	// Non-positive restores the default rather than disabling detection.
	SetStaleGoroutineThreshold(0)
	if got := GetStaleGoroutineThreshold(); got != defaultStaleThreshold {
		t.Errorf("threshold after 0 = %v, want default %v", got, defaultStaleThreshold)
	}
	SetStaleGoroutineThreshold(-5 * time.Hour)
	if got := GetStaleGoroutineThreshold(); got != defaultStaleThreshold {
		t.Errorf("threshold after negative = %v, want default %v", got, defaultStaleThreshold)
	}
}

func TestStaleDetection(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()
	SetStaleGoroutineThreshold(60 * time.Minute)

	report := leakReportFor([]string{
		block(1, "chan receive", 59, "justUnder"), // below threshold
		block(2, "chan receive", 60, "atExactly"), // at threshold, inclusive
		block(3, "select", 600, "wayOver"),
		block(4, "running", 0, "healthy"),
	})

	if report.StaleGoroutines != 2 {
		t.Errorf("stale goroutines = %d, want 2", report.StaleGoroutines)
	}
	if !report.LeakSuspected {
		t.Error("expected LeakSuspected with stale goroutines present")
	}
	if report.StaleThresholdMinutes != 60 {
		t.Errorf("threshold minutes = %d, want 60", report.StaleThresholdMinutes)
	}
	if len(report.SuspiciousGroups) != 2 {
		t.Fatalf("suspicious groups = %d, want 2", len(report.SuspiciousGroups))
	}
	// Worst first: the 600-minute group leads.
	if report.SuspiciousGroups[0].BlockedMinutes != 600 {
		t.Errorf("first suspicious group blocked = %d, want 600",
			report.SuspiciousGroups[0].BlockedMinutes)
	}
	if !strings.Contains(report.Message, "Possible goroutine leak") {
		t.Errorf("unexpected message: %q", report.Message)
	}
}

func TestNoLeakReportedWhenHealthy(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	report := leakReportFor([]string{
		block(1, "running", 0, "a"),
		block(2, "IO wait", 3, "b"),
	})

	if report.LeakSuspected {
		t.Errorf("expected no leak, got message: %q", report.Message)
	}
	if report.StaleGoroutines != 0 || report.GrowingGroups != 0 {
		t.Errorf("stale=%d growing=%d, want 0/0", report.StaleGoroutines, report.GrowingGroups)
	}
	if len(report.SuspiciousGroups) != 0 {
		t.Errorf("expected no suspicious groups, got %d", len(report.SuspiciousGroups))
	}
	if !strings.Contains(report.Message, "No goroutine leak detected") {
		t.Errorf("unexpected message: %q", report.Message)
	}
}

// Growth must not be reported until the snapshot window is full, otherwise a
// service that has just started reports a leak from two data points.
func TestGrowthRequiresFullSnapshotWindow(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	// Rising every time, but one snapshot short of the window.
	for i := 1; i < defaultSnapshotWindow; i++ {
		blocks := make([]string, 0, i)
		for j := 0; j < i; j++ {
			blocks = append(blocks, block(j+1, "chan receive", 0, "leaky"))
		}
		recordSnapshot(groupGoroutines(blocks))
	}

	report := leakReportFor([]string{block(1, "chan receive", 0, "leaky")})
	if report.GrowingGroups != 0 {
		t.Errorf("growing groups = %d, want 0 before the window is full", report.GrowingGroups)
	}
	if report.SnapshotsRetained >= report.SnapshotsRequired {
		t.Errorf("retained %d of %d; expected an incomplete window",
			report.SnapshotsRetained, report.SnapshotsRequired)
	}
	if !strings.Contains(report.Message, "warming up") {
		t.Errorf("expected a warming-up message, got %q", report.Message)
	}
}

// A stack whose count rises at every snapshot across a full window is a leak.
func TestGrowthDetectedForMonotonicIncrease(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	var final []string
	for i := 1; i <= defaultSnapshotWindow; i++ {
		blocks := make([]string, 0, i)
		for j := 0; j < i; j++ {
			blocks = append(blocks, block(j+1, "chan receive", 0, "leaky"))
		}
		recordSnapshot(groupGoroutines(blocks))
		final = blocks
	}

	report := leakReportFor(final)
	if report.GrowingGroups != 1 {
		t.Fatalf("growing groups = %d, want 1", report.GrowingGroups)
	}
	if !report.LeakSuspected {
		t.Error("expected LeakSuspected for a monotonically growing stack")
	}

	g := report.SuspiciousGroups[0]
	if !g.Growing {
		t.Error("expected the group to be marked Growing")
	}
	// From 1 in the oldest retained snapshot to defaultSnapshotWindow in the newest.
	if want := defaultSnapshotWindow - 1; g.Growth != want {
		t.Errorf("growth = %d, want %d", g.Growth, want)
	}
	if !strings.Contains(report.Message, "growing across the last") {
		t.Errorf("unexpected message: %q", report.Message)
	}
}

// A pool that scales up and back down, or holds steady, is not a leak.
func TestGrowthNotDetectedForNonMonotonicCounts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		counts []int
	}{
		{"steady", []int{3, 3, 3, 3, 3}},
		{"spike then settle", []int{1, 5, 5, 2, 2}},
		{"sawtooth", []int{1, 2, 1, 2, 3}},
		{"shrinking", []int{5, 4, 3, 2, 1}},
		{"plateau at the end", []int{1, 2, 3, 4, 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(ResetLeakDetectionState)
			ResetLeakDetectionState()

			var final []string
			for _, n := range tc.counts {
				blocks := make([]string, 0, n)
				for j := 0; j < n; j++ {
					blocks = append(blocks, block(j+1, "chan receive", 0, "pool"))
				}
				recordSnapshot(groupGoroutines(blocks))
				final = blocks
			}

			report := leakReportFor(final)
			if report.GrowingGroups != 0 {
				t.Errorf("counts %v: growing groups = %d, want 0", tc.counts, report.GrowingGroups)
			}
		})
	}
}

// A stack absent from the oldest snapshot is new, not growing -- otherwise
// every short-lived request handler is flagged.
func TestGrowthNotDetectedForNewlyAppearedStack(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	for i := 0; i < defaultSnapshotWindow; i++ {
		recordSnapshot(groupGoroutines([]string{block(1, "running", 0, "existing")}))
	}

	report := leakReportFor([]string{
		block(1, "running", 0, "existing"),
		block(2, "chan receive", 0, "brandNew"),
	})
	if report.GrowingGroups != 0 {
		t.Errorf("growing groups = %d, want 0 for a stack with no baseline", report.GrowingGroups)
	}
}

// The retained window must not grow without bound.
func TestSnapshotWindowIsBounded(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	for i := 0; i < defaultSnapshotWindow*10; i++ {
		recordSnapshot(groupGoroutines([]string{block(1, "running", 0, "x")}))
	}

	leakMu.RLock()
	retained := len(snapshots)
	leakMu.RUnlock()

	if retained != defaultSnapshotWindow {
		t.Errorf("retained %d snapshots, want %d", retained, defaultSnapshotWindow)
	}
}

// The API response must stay bounded no matter how pathological the process.
func TestSuspiciousGroupsAreCapped(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()
	SetStaleGoroutineThreshold(time.Minute)

	blocks := make([]string, 0, maxSuspiciousGroups*3)
	for i := 0; i < maxSuspiciousGroups*3; i++ {
		blocks = append(blocks, block(i+1, "chan receive", 60+i, fmt.Sprintf("stack%d", i)))
	}

	report := leakReportFor(blocks)
	if len(report.SuspiciousGroups) != maxSuspiciousGroups {
		t.Errorf("suspicious groups = %d, want the cap of %d", len(report.SuspiciousGroups), maxSuspiciousGroups)
	}
	// Totals must still reflect everything, not just what was returned.
	if report.StaleGoroutines != maxSuspiciousGroups*3 {
		t.Errorf("stale goroutines = %d, want %d", report.StaleGoroutines, maxSuspiciousGroups*3)
	}
}

// Ordering must be deterministic so the dashboard does not reshuffle on refresh.
func TestSuspiciousGroupOrderingIsDeterministic(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()
	SetStaleGoroutineThreshold(time.Minute)

	blocks := []string{
		block(1, "chan receive", 10, "a"),
		block(2, "select", 500, "b"),
		block(3, "semacquire", 100, "c"),
	}

	var first []string
	for i := 0; i < 20; i++ {
		report := leakReportFor(blocks)
		order := make([]string, 0, len(report.SuspiciousGroups))
		for _, g := range report.SuspiciousGroups {
			order = append(order, g.Signature)
		}
		if i == 0 {
			first = order
			continue
		}
		if strings.Join(order, ",") != strings.Join(first, ",") {
			t.Fatalf("ordering changed between passes: %v vs %v", first, order)
		}
	}
	// Longest-blocked leads.
	if got := leakReportFor(blocks).SuspiciousGroups[0].BlockedMinutes; got != 500 {
		t.Errorf("first group blocked = %d, want 500", got)
	}
}

// Signatures identify a call stack, so they must be stable across passes and
// distinct between different stacks.
func TestSignatureStability(t *testing.T) {
	a := callStackOf(block(1, "chan receive", 0, "worker"))
	b := callStackOf(block(99, "select", 4000, "worker")) // same stack, different header
	c := callStackOf(block(1, "chan receive", 0, "other"))

	if signatureOf(a) != signatureOf(b) {
		t.Error("expected identical signatures for the same call stack")
	}
	if signatureOf(a) == signatureOf(c) {
		t.Error("expected different signatures for different call stacks")
	}
	if signatureOf(a) != signatureOf(a) {
		t.Error("signature is not deterministic")
	}
}

// AnalyzeGoroutineLeaks is called from the metrics path; concurrent calls must
// be safe. Run under -race.
func TestAnalyzeGoroutineLeaksConcurrent(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); AnalyzeGoroutineLeaks() }()
		go func() { defer wg.Done(); _ = GetStaleGoroutineThreshold() }()
		go func() { defer wg.Done(); SetStaleGoroutineThreshold(time.Hour) }()
	}
	wg.Wait()
}

// A real pass over the live process must produce a coherent report.
func TestAnalyzeGoroutineLeaksAgainstLiveProcess(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	report := AnalyzeGoroutineLeaks()
	if report == nil {
		t.Fatal("expected a report")
	}
	if report.TotalGoroutines <= 0 {
		t.Errorf("total goroutines = %d, want > 0", report.TotalGoroutines)
	}
	if report.SnapshotsRetained != 1 {
		t.Errorf("snapshots retained = %d, want 1 after one pass", report.SnapshotsRetained)
	}
	// The test binary has not been running for 24h, so nothing can be stale.
	if report.StaleGoroutines != 0 {
		t.Errorf("stale goroutines = %d, want 0 in a fresh process", report.StaleGoroutines)
	}
	if report.Message == "" {
		t.Error("expected a non-empty message")
	}
}

func TestFormatThreshold(t *testing.T) {
	for _, tc := range []struct {
		in   time.Duration
		want string
	}{
		{24 * time.Hour, "24h"},
		{48 * time.Hour, "48h"},
		{90 * time.Minute, "90m"},
		{25 * time.Hour, "25h"},
		{time.Hour, "1h"},
		{30 * time.Minute, "30m"},
		{time.Minute, "1m"},
	} {
		if got := formatThreshold(tc.in); got != tc.want {
			t.Errorf("formatThreshold(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// End-to-end: leak real goroutines in this process and confirm the detector
// sees them growing. This exercises the whole path -- runtime.Stack capture,
// splitting, header parsing, grouping, snapshotting, growth analysis -- against
// genuine runtime output rather than fixtures.
func TestDetectsRealLeakingGoroutines(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	// Goroutines parked on a channel nobody will ever send to: the classic leak.
	park := make(chan struct{})
	done := make(chan struct{})
	var spawned sync.WaitGroup

	leak := func(n int) {
		for i := 0; i < n; i++ {
			spawned.Add(1)
			go func() {
				defer spawned.Done()
				select {
				case <-park:
				case <-done:
				}
			}()
		}
		// Let the runtime actually park them before capturing.
		time.Sleep(50 * time.Millisecond)
	}

	// Release every leaked goroutine on the way out, so this test does not
	// itself leak into the rest of the package's tests.
	t.Cleanup(func() {
		close(done)
		spawned.Wait()
	})

	// Leak before every snapshot, one more time than the window holds, so the
	// retained window is full AND every step in it rose. A trailing analysis
	// with no new goroutines would make the final step flat, which the
	// monotonic rule correctly refuses to call a leak.
	var report *models.GoroutineLeakReport
	for i := 0; i < defaultSnapshotWindow+1; i++ {
		leak(5)
		report = AnalyzeGoroutineLeaks()
	}
	if report.GrowingGroups == 0 {
		t.Fatalf("expected a growing group; report: %+v", report.Message)
	}
	if !report.LeakSuspected {
		t.Error("expected LeakSuspected for genuinely leaking goroutines")
	}

	// The growing group should be the parked select, and its growth should
	// reflect the goroutines added across the retained window.
	var found bool
	for _, g := range report.SuspiciousGroups {
		if g.Growing && strings.Contains(g.CallStack, "TestDetectsRealLeakingGoroutines") {
			found = true
			if g.Growth <= 0 {
				t.Errorf("growth = %d, want > 0", g.Growth)
			}
			if g.Count < defaultSnapshotWindow*5 {
				t.Errorf("count = %d, want at least %d", g.Count, defaultSnapshotWindow*5)
			}
		}
	}
	if !found {
		var stacks []string
		for _, g := range report.SuspiciousGroups {
			stacks = append(stacks, g.State)
		}
		t.Errorf("did not find the leaking stack among suspicious groups (states: %v)", stacks)
	}
}

// A process whose goroutine count is stable must not be reported as leaking,
// even over a full window. Guards against the detector crying wolf.
func TestStableProcessIsNotReportedAsLeaking(t *testing.T) {
	t.Cleanup(ResetLeakDetectionState)
	ResetLeakDetectionState()

	for i := 0; i < defaultSnapshotWindow+2; i++ {
		AnalyzeGoroutineLeaks()
		time.Sleep(10 * time.Millisecond)
	}

	report := AnalyzeGoroutineLeaks()
	if report.LeakSuspected {
		t.Errorf("stable process reported as leaking: %s", report.Message)
	}
}

// runtime.Stack truncates silently and reports only bytes written, so the only
// signal of truncation is a result that exactly fills the buffer. These cases
// cover the retry logic that replaced the old fixed 1 MB single call.
func TestCaptureStacksGrowsUntilTheDumpFits(t *testing.T) {
	const payloadSize = 5000

	t.Run("fits on the first attempt", func(t *testing.T) {
		calls := 0
		dump := func(buf []byte, all bool) int {
			calls++
			return copy(buf, strings.Repeat("x", 10))
		}
		got := captureStacks(dump, 1024, 1<<20)
		if len(got) != 10 {
			t.Errorf("len = %d, want 10", len(got))
		}
		if calls != 1 {
			t.Errorf("dump called %d times, want 1", calls)
		}
	})

	t.Run("retries with a larger buffer until it fits", func(t *testing.T) {
		calls := 0
		sizes := []int{}
		dump := func(buf []byte, all bool) int {
			calls++
			sizes = append(sizes, len(buf))
			// Simulate a dump of payloadSize bytes: fills (and so appears
			// truncated) for any buffer that cannot hold it.
			return copy(buf, strings.Repeat("x", payloadSize))
		}

		got := captureStacks(dump, 1024, 1<<20)
		if len(got) != payloadSize {
			t.Errorf("len = %d, want the full %d bytes", len(got), payloadSize)
		}
		// 1024 -> 2048 -> 4096 -> 8192, the first that exceeds payloadSize.
		want := []int{1024, 2048, 4096, 8192}
		if len(sizes) != len(want) {
			t.Fatalf("buffer sizes tried = %v, want %v", sizes, want)
		}
		for i := range want {
			if sizes[i] != want[i] {
				t.Errorf("attempt %d used %d bytes, want %d", i+1, sizes[i], want[i])
			}
		}
	})

	t.Run("stops doubling at the cap and accepts truncation", func(t *testing.T) {
		calls := 0
		dump := func(buf []byte, all bool) int {
			calls++
			// Never fits, whatever the buffer size.
			for i := range buf {
				buf[i] = 'x'
			}
			return len(buf)
		}

		got := captureStacks(dump, 1024, 4096)
		if len(got) != 4096 {
			t.Errorf("len = %d, want the capped %d", len(got), 4096)
		}
		// 1024, 2048, 4096 -- then the cap is reached and it gives up.
		if calls != 3 {
			t.Errorf("dump called %d times, want 3 before hitting the cap", calls)
		}
	})

	t.Run("a truncating fixed buffer is what the old code did", func(t *testing.T) {
		// Regression guard: with a single fixed call the payload would be
		// clipped to the buffer size, undercounting goroutines.
		dump := func(buf []byte, all bool) int {
			return copy(buf, strings.Repeat("x", payloadSize))
		}
		clipped := captureStacks(dump, 1024, 1024) // cap == initial: no retry possible
		if len(clipped) != 1024 {
			t.Errorf("len = %d, want the clipped %d", len(clipped), 1024)
		}
		// And with retry enabled the same dump comes back whole.
		full := captureStacks(dump, 1024, 1<<20)
		if len(full) != payloadSize {
			t.Errorf("len = %d, want the full %d", len(full), payloadSize)
		}
	})
}

// captureAllStacks must return a complete dump for the live process.
func TestCaptureAllStacksReturnsCompleteDump(t *testing.T) {
	got := captureAllStacks()
	if got == "" {
		t.Fatal("expected a non-empty stack dump")
	}
	if !strings.HasPrefix(got, "goroutine ") {
		t.Errorf("dump does not start with a goroutine header: %.60q", got)
	}
	// A complete dump ends with a full block, not mid-line.
	if !strings.HasSuffix(got, "\n") {
		t.Error("dump does not end on a line boundary, suggesting truncation")
	}
	// Every block must be parsable, which a clipped final block would not be.
	blocks := SplitGoroutines(got)
	for i, b := range blocks {
		if _, _, _, ok := parseGoroutineHeader(b); !ok {
			t.Errorf("block %d of %d is not parsable: %.80q", i, len(blocks), b)
		}
	}
}

func TestCallStackOfHandlesHeaderOnlyBlock(t *testing.T) {
	if got := callStackOf("goroutine 1 [running]:"); got != "" {
		t.Errorf("call stack for a header-only block = %q, want empty", got)
	}
	if got := callStackOf(""); got != "" {
		t.Errorf("call stack for an empty block = %q, want empty", got)
	}
	got := callStackOf("goroutine 1 [running]:\nmain.main()\n\t/app/main.go:9 +0x1c\n")
	if strings.Contains(got, "goroutine 1") {
		t.Errorf("call stack still contains the header: %q", got)
	}
	if !strings.Contains(got, "main.main()") {
		t.Errorf("call stack lost its body: %q", got)
	}
}
