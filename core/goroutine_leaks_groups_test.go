package core

import (
	"strings"
	"testing"
)

// The dashboard breaks goroutines down by state and ranks them by blocked
// time, which needs every group -- a report carrying only the offenders cannot
// say what the non-suspicious goroutines are doing. These guard the contract
// the UI depends on.

func TestReportCarriesEveryGroupNotJustSuspicious(t *testing.T) {
	ResetLeakDetectionState()

	report := AnalyzeGoroutineLeaks()
	if report == nil {
		t.Fatal("AnalyzeGoroutineLeaks returned nil")
	}

	if len(report.Groups) == 0 {
		t.Fatal("Groups is empty; the test binary always has running goroutines")
	}
	if report.GroupsTotal < len(report.Groups) {
		t.Errorf("GroupsTotal %d is below the %d groups returned; it must count "+
			"what exists, not what survived the cap",
			report.GroupsTotal, len(report.Groups))
	}
	if len(report.Groups) > maxGroups {
		t.Errorf("Groups has %d entries, above the cap of %d", len(report.Groups), maxGroups)
	}

	// Groups is a superset of SuspiciousGroups.
	inAll := map[string]bool{}
	for _, g := range report.Groups {
		inAll[g.Signature] = true
	}
	for _, s := range report.SuspiciousGroups {
		if !inAll[s.Signature] {
			t.Errorf("suspicious group %s is missing from Groups", s.Signature)
		}
	}

	// Counts across all groups must reconcile with the total.
	sum := 0
	for _, g := range report.Groups {
		sum += g.Count
	}
	if report.GroupsTotal <= maxGroups && sum != report.TotalGoroutines {
		t.Errorf("group counts sum to %d but TotalGoroutines is %d",
			sum, report.TotalGoroutines)
	}

	for _, g := range report.Groups {
		if g.Count <= 0 {
			t.Errorf("group %s has non-positive count %d", g.Signature, g.Count)
		}
		if strings.TrimSpace(g.State) == "" {
			t.Errorf("group %s has an empty state; the UI groups by it", g.Signature)
		}
	}
}

func TestGroupsAreOrderedWorstFirst(t *testing.T) {
	ResetLeakDetectionState()

	report := AnalyzeGoroutineLeaks()
	for i := 1; i < len(report.Groups); i++ {
		a, b := report.Groups[i-1], report.Groups[i]
		switch {
		case a.BlockedMinutes != b.BlockedMinutes:
			if a.BlockedMinutes < b.BlockedMinutes {
				t.Fatalf("group %d blocked %dm sorts before %d blocked %dm",
					i-1, a.BlockedMinutes, i, b.BlockedMinutes)
			}
		case a.Growth != b.Growth:
			if a.Growth < b.Growth {
				t.Fatalf("group %d growth %d sorts before %d growth %d",
					i-1, a.Growth, i, b.Growth)
			}
		case a.Count != b.Count:
			if a.Count < b.Count {
				t.Fatalf("group %d count %d sorts before %d count %d",
					i-1, a.Count, i, b.Count)
			}
		default:
			if a.Signature > b.Signature {
				t.Fatalf("equal groups are not in signature order: %s before %s",
					a.Signature, b.Signature)
			}
		}
	}
}
