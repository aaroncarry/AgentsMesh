package slugkit

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

// TestReservedListSyncAcrossSources ensures the backend reserved list stays
// in lock-step with two sibling hardcoded sources:
//
//   - web/src/lib/slug/reserved.ts  (frontend Validate)
//   - backend/migrations/000115_add_slug_reserved_check.up.sql  (DB CHECK)
//
// Drift causes user-facing bugs: a slug allowed by one tier and blocked by
// another. Adding/removing a reserved word means editing all three sources.
//
// This test fails (not skips) when sibling files cannot be located.
// Set SLUGKIT_SKIP_FRONTEND_SYNC_CHECK=1 to opt out in environments that
// genuinely cannot see them (e.g. backend-only image builds).
func TestReservedListSyncAcrossSources(t *testing.T) {
	if os.Getenv("SLUGKIT_SKIP_FRONTEND_SYNC_CHECK") == "1" {
		t.Skip("cross-source sync check disabled by SLUGKIT_SKIP_FRONTEND_SYNC_CHECK=1")
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("could not locate repo root (set SLUGKIT_SKIP_FRONTEND_SYNC_CHECK=1 to bypass): %v", err)
	}

	tsSet := readReservedFromFile(t, filepath.Join(repoRoot, "web", "src", "lib", "slug", "reserved.ts"), `"([a-z0-9-]+)"`)
	sqlPath := latestReservedMigration(t, filepath.Join(repoRoot, "backend", "migrations"))
	sqlSet := readReservedFromFile(t, sqlPath, `'([a-z0-9-]+)'`)

	goSet := make(map[string]bool, len(reserved))
	for k := range reserved {
		goSet[k] = true
	}

	assertSetsEqual(t, "go", goSet, "ts", tsSet)
	assertSetsEqual(t, "go", goSet, "sql", sqlSet)
}

func readReservedFromFile(t *testing.T, path, pattern string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	out := extractReservedSet(string(data), pattern)
	if len(out) == 0 {
		t.Fatalf("no reserved entries parsed from %s (pattern=%s)", path, pattern)
	}
	return out
}

func extractReservedSet(src, pattern string) map[string]bool {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(src, -1)
	out := make(map[string]bool, len(matches))
	for _, m := range matches {
		out[m[1]] = true
	}
	return out
}

func assertSetsEqual(t *testing.T, leftName string, left map[string]bool, rightName string, right map[string]bool) {
	t.Helper()
	for word := range left {
		if !right[word] {
			t.Errorf("%s reserved %q is missing from %s", leftName, word, rightName)
		}
	}
	for word := range right {
		if !left[word] {
			t.Errorf("%s reserved %q is missing from %s", rightName, word, leftName)
		}
	}
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			parent := filepath.Dir(dir)
			if _, err := os.Stat(filepath.Join(parent, "web")); err == nil {
				return parent, nil
			}
		}
		dir = filepath.Dir(dir)
	}
	return "", os.ErrNotExist
}

// latestReservedMigration returns the highest-numbered migration matching
// *slug_reserved*.up.sql. The pattern is intentionally narrow so business
// migrations whose names happen to contain "reserved" don't get mis-detected
// as the SSOT for reserved-slug check constraints.
func latestReservedMigration(t *testing.T, migrationsDir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(migrationsDir, "*slug_reserved*.up.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no *slug_reserved*.up.sql migration found under %s", migrationsDir)
	}
	sort.Strings(matches) // lexicographic = numeric since migrations are zero-padded
	return matches[len(matches)-1]
}

// TestReservedListIsSorted ensures the canonical list is deterministic so
// drift diffs stay readable.
func TestReservedListIsSorted(t *testing.T) {
	list := ReservedList()
	sort.Strings(list)
	for i := 1; i < len(list); i++ {
		if list[i] == list[i-1] {
			t.Errorf("duplicate entry: %q", list[i])
		}
	}
}
