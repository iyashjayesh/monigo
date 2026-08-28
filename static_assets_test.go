package monigo

import (
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The dashboard is compiled into the consuming service's binary via
// //go:embed static/*, served offline, and has no build step, linter, or test
// runner of its own. These tests are the only automated verification the UI
// gets, so they guard the three failure modes that have actually happened:
// a reference to a file that is not there, a dependency fetched from the
// internet, and unreferenced weight shipping to every user.

// resourceRefPattern matches references that the browser must FETCH for the page
// to work: stylesheets, scripts and images. Anchor hrefs are deliberately
// excluded -- an <a> pointing at github.com is a link, not a dependency.
var resourceRefPattern = regexp.MustCompile(`(?is)<(?:link|script|img)\b[^>]*?\b(?:href|src)\s*=\s*"([^"]*)"`)

// cssURLPattern matches url(...) in stylesheets, which are fetched the same way.
var cssURLPattern = regexp.MustCompile(`(?i)url\(\s*['"]?([^'")]+)['"]?\s*\)`)

// htmlCommentPattern and cssCommentPattern match commented-out regions. These
// must be stripped before scanning: a commented <script> tag fetches nothing, so
// treating one as a live reference produces a false positive. Three commented-out
// references to a non-existent js/core/main.js exist today and are NOT defects.
var htmlCommentPattern = regexp.MustCompile(`(?s)<!--.*?-->`)
var cssCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)

// stripComments removes commented-out regions so only live references are scanned.
func stripComments(content, ext string) string {
	if strings.EqualFold(ext, ".css") {
		return cssCommentPattern.ReplaceAllString(content, "")
	}
	return htmlCommentPattern.ReplaceAllString(content, "")
}

// knownExternalResources are remote fetches the dashboard is permitted to make.
//
// It is empty, and should stay that way: the dashboard is served from inside a
// consuming service's binary and must render on an airgapped host. Anything
// added here is a page that breaks on a machine with no route to the internet.
//
// It previously held three entries -- Font Awesome and html2canvas from cdnjs,
// and a Lato webfont the vendored template CSS pulled from Google Fonts. All
// three were removed when the inline icon sprite replaced Font Awesome.
var knownExternalResources = map[string]string{}

// maxEmbeddedBytes caps the total size of the embedded dashboard. Every byte
// ships inside every consuming binary and is downloaded by every `go get`, so
// this is a real cost paid by users who may never open the dashboard.
//
// Tighten this whenever the payload legitimately shrinks. Do not raise it
// without a note in the PR saying what was added and why it earns its size.
//
// Headroom over the current payload is deliberately small (~5%), so that
// anything substantial being added has to be an explicit decision.
const maxEmbeddedBytes = 8_300_000

// htmlFiles returns every embedded .html file.
func htmlFiles(t *testing.T) []string {
	t.Helper()
	var out []string
	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(path.Ext(p), ".html") {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded files: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("no embedded .html files found; the embed directive may have changed")
	}
	return out
}

// isRemote reports whether a reference points off-host.
func isRemote(ref string) bool {
	return strings.HasPrefix(ref, "http://") ||
		strings.HasPrefix(ref, "https://") ||
		strings.HasPrefix(ref, "//")
}

// isFetchable reports whether a reference names a file that has to exist.
func isFetchable(ref string) bool {
	if ref == "" || ref == "#" {
		return false
	}
	for _, skip := range []string{"#", "data:", "mailto:", "tel:", "javascript:", "about:"} {
		if strings.HasPrefix(ref, skip) {
			return false
		}
	}
	return !isRemote(ref)
}

// resolve turns a reference as written in a file into an embedded FS path.
// Query strings and fragments are stripped; "./x" and "../x" are resolved
// relative to the referencing file's directory.
func resolve(fromFile, ref string) (string, error) {
	if u, err := url.Parse(ref); err == nil {
		ref = u.Path
	}
	if ref == "" {
		return "", fmt.Errorf("empty path after stripping query")
	}
	joined := path.Join(path.Dir(fromFile), ref)
	if !strings.HasPrefix(joined, "static/") {
		return joined, fmt.Errorf("resolves outside the embedded tree")
	}
	return joined, nil
}

// Every asset an embedded page tells the browser to fetch must exist in the
// embedded filesystem. A missing one is a 404 on every page load, invisible
// unless someone opens devtools -- which is how js/core/main.js survived.
func TestEmbeddedPagesReferenceOnlyExistingAssets(t *testing.T) {
	type miss struct{ file, ref, resolved, why string }
	var misses []miss

	for _, f := range htmlFiles(t) {
		b, err := fs.ReadFile(staticFiles, f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		content := stripComments(string(b), path.Ext(f))
		for _, m := range resourceRefPattern.FindAllStringSubmatch(content, -1) {
			ref := strings.TrimSpace(m[1])
			if !isFetchable(ref) {
				continue
			}
			resolved, err := resolve(f, ref)
			if err != nil {
				misses = append(misses, miss{f, ref, resolved, err.Error()})
				continue
			}
			if _, err := fs.Stat(staticFiles, resolved); err != nil {
				misses = append(misses, miss{f, ref, resolved, "not present in embed.FS"})
			}
		}
	}

	if len(misses) > 0 {
		sort.Slice(misses, func(i, j int) bool {
			if misses[i].file != misses[j].file {
				return misses[i].file < misses[j].file
			}
			return misses[i].ref < misses[j].ref
		})
		var b strings.Builder
		fmt.Fprintf(&b, "%d referenced asset(s) missing from the embedded dashboard:\n", len(misses))
		for _, m := range misses {
			fmt.Fprintf(&b, "  %s\n    references %q\n    -> %s (%s)\n", m.file, m.ref, m.resolved, m.why)
		}
		b.WriteString("\nEach of these is a failed request on every load of that page.")
		t.Error(b.String())
	}
}

// The dashboard must work with no internet access. This test does not ban
// remote resources outright -- three exist today -- but it pins the set, so
// adding a fourth is a deliberate act rather than an accident.
func TestEmbeddedPagesAddNoNewExternalResources(t *testing.T) {
	files := htmlFiles(t)

	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(path.Ext(p), ".css") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded files: %v", err)
	}

	allowed := func(ref string) bool {
		for prefix := range knownExternalResources {
			if strings.HasPrefix(ref, prefix) {
				return true
			}
		}
		return false
	}

	seen := map[string][]string{}
	for _, f := range files {
		b, err := fs.ReadFile(staticFiles, f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		content := stripComments(string(b), path.Ext(f))

		refs := []string{}
		for _, m := range resourceRefPattern.FindAllStringSubmatch(content, -1) {
			refs = append(refs, strings.TrimSpace(m[1]))
		}
		if strings.EqualFold(path.Ext(f), ".css") {
			for _, m := range cssURLPattern.FindAllStringSubmatch(content, -1) {
				refs = append(refs, strings.TrimSpace(m[1]))
			}
		}

		for _, ref := range refs {
			if !isRemote(ref) || allowed(ref) {
				continue
			}
			seen[ref] = append(seen[ref], f)
		}
	}

	if len(seen) > 0 {
		keys := make([]string, 0, len(seen))
		for k := range seen {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var b strings.Builder
		b.WriteString("new external resource dependencies found in the embedded dashboard:\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s\n    in: %s\n", k, strings.Join(seen[k], ", "))
		}
		b.WriteString("\nThe dashboard must render on an airgapped host. Vendor the asset, or\n")
		b.WriteString("add it to knownExternalResources with a reason and a plan to remove it.")
		t.Error(b.String())
	}
}

// Known remote resources must actually still be present. Without this, the
// allowlist above silently rots into a list of things that were removed years
// ago, and stops protecting anything.
func TestKnownExternalResourcesAreStillPresent(t *testing.T) {
	var content strings.Builder
	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(path.Ext(p))
		if ext != ".html" && ext != ".css" {
			return nil
		}
		b, err := fs.ReadFile(staticFiles, p)
		if err != nil {
			return err
		}
		content.WriteString(stripComments(string(b), ext))
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded files: %v", err)
	}

	for ref, reason := range knownExternalResources {
		if !strings.Contains(content.String(), ref) {
			t.Errorf("knownExternalResources lists %q (%s) but it is no longer referenced;\n"+
				"remove the entry so the allowlist keeps shrinking", ref, reason)
		}
	}
}

// The embedded dashboard ships inside every consuming binary, so its size is a
// cost borne by users who may never open it.
func TestEmbeddedPayloadWithinBudget(t *testing.T) {
	var total int64
	sizes := map[string]int64{}

	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		sizes[p] = info.Size()
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded files: %v", err)
	}

	t.Logf("embedded dashboard: %.1f MB across %d files (budget %.1f MB)",
		float64(total)/1e6, len(sizes), float64(maxEmbeddedBytes)/1e6)

	if total > maxEmbeddedBytes {
		type entry struct {
			p string
			n int64
		}
		list := make([]entry, 0, len(sizes))
		for p, n := range sizes {
			list = append(list, entry{p, n})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })

		var b strings.Builder
		fmt.Fprintf(&b, "embedded dashboard is %d bytes, over the %d byte budget by %d.\n",
			total, maxEmbeddedBytes, total-maxEmbeddedBytes)
		b.WriteString("largest contributors:\n")
		for i, e := range list {
			if i >= 8 {
				break
			}
			fmt.Fprintf(&b, "  %8.2f MB  %s\n", float64(e.n)/1e6, e.p)
		}
		t.Error(b.String())
	}
}

// Editor and OS droppings should not be compiled into users' binaries.
func TestEmbeddedPayloadHasNoJunkFiles(t *testing.T) {
	junk := []string{".DS_Store", "Thumbs.db", ".gitkeep"}
	var found []string

	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		for _, j := range junk {
			if strings.EqualFold(path.Base(p), j) {
				found = append(found, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded files: %v", err)
	}

	if len(found) > 0 {
		sort.Strings(found)
		t.Errorf("junk files embedded into the binary:\n  %s\n\nDelete them and add the pattern to .gitignore.",
			strings.Join(found, "\n  "))
	}
}

// A CSS custom property that is never defined does not error -- the browser
// drops the declaration and the element silently keeps its inherited value.
// That is exactly how a footer shipped with var(--bg)/var(--dim2), names taken
// from the design file rather than from this stylesheet, and rendered with a
// transparent background and the wrong text tone while looking plausible.
//
// var(--x, fallback) is legitimate and exempt: the author has said what to do
// when the token is absent.
func TestStylesheetsDefineEveryTokenTheyUse(t *testing.T) {
	comments := regexp.MustCompile(`(?s)/\*.*?\*/`)
	defRe := regexp.MustCompile(`(--[\w-]+)\s*:`)
	// Only var(--x) and var( --x ) -- a comma means a fallback was supplied.
	useRe := regexp.MustCompile(`var\(\s*(--[\w-]+)\s*\)`)

	// Only MoniGo's own stylesheet. Everything under static/css/core is
	// vendored Bootstrap and template CSS: backend.css has a genuinely dead
	// var(--fill-percentage) that no rule, script, or page ever sets, and
	// rewriting 26k lines of third-party CSS is not this test's job.
	const ours = "static/css/monigo-styles.css"

	err := fs.WalkDir(staticFiles, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || p != ours {
			return err
		}
		raw, err := staticFiles.ReadFile(p)
		if err != nil {
			return err
		}
		css := comments.ReplaceAllString(string(raw), "")

		defined := map[string]bool{}
		for _, m := range defRe.FindAllStringSubmatch(css, -1) {
			defined[m[1]] = true
		}

		var missing []string
		seen := map[string]bool{}
		for _, m := range useRe.FindAllStringSubmatch(css, -1) {
			if !defined[m[1]] && !seen[m[1]] {
				seen[m[1]] = true
				missing = append(missing, m[1])
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("%s uses custom properties it never defines: %s\n"+
				"These resolve to nothing and the declaration is dropped silently. "+
				"Either define them, use the name this stylesheet actually declares, "+
				"or supply a fallback as var(%s, <value>).",
				p, strings.Join(missing, ", "), missing[0])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking embedded stylesheets: %v", err)
	}
	if _, err := staticFiles.ReadFile(ours); err != nil {
		t.Fatalf("%s is gone, so this test silently checks nothing: %v", ours, err)
	}
}
