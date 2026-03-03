//#:[.'.]:>-==================================================================================
//#:[.'.]:>- Marco Antonio - markitos devsecops kulture
//#:[.'.]:>- The Way of the Artisan
//#:[.'.]:>- markitos.es.info@gmail.com
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-it/repositories
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-public/repositories
//#:[.'.]:>- 📺 https://www.youtube.com/@markitos_devsecops
//#:[.'.]:>- =================================================================================

package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteChangelog(t *testing.T) {
	dir := t.TempDir()

	result := ReleaseResult{
		Previous: "v1.0.0",
		Next:     "v1.1.0",
		Reason:   "feat detectado → bump MINOR",
		Commits: []Commit{
			{Hash: "abc1234", Type: "feat", Scope: "auth", Subject: "add oauth2"},
			{Hash: "def5678", Type: "fix", Scope: "", Subject: "fix null pointer"},
		},
	}

	if err := writeChangelog(dir, result); err != nil {
		t.Fatalf("writeChangelog: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	mustContain(t, content, "## v1.1.0")
	mustContain(t, content, "✨ Features")
	mustContain(t, content, "🐛 Bug Fixes")
	mustContain(t, content, "add oauth2")
	mustContain(t, content, "fix null pointer")
	mustContain(t, content, "`abc1234`")
	mustContain(t, content, "(auth)")
}

func TestWriteChangelogPrepends(t *testing.T) {
	dir := t.TempDir()
	existing := "## v1.0.0 (2024-01-01)\n\n### 🐛 Bug Fixes\n\n- old fix\n\n"
	os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(existing), 0644)

	result := ReleaseResult{
		Next:    "v1.1.0",
		Commits: []Commit{{Hash: "abc1234", Type: "feat", Subject: "new feature"}},
	}

	if err := writeChangelog(dir, result); err != nil {
		t.Fatalf("writeChangelog: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	content := string(data)

	mustContain(t, content, "## v1.1.0")
	mustContain(t, content, "## v1.0.0")

	if strings.Index(content, "v1.1.0") > strings.Index(content, "v1.0.0") {
		t.Error("v1.1.0 should appear before v1.0.0")
	}
}

func TestWriteChangelogBreakingEntry(t *testing.T) {
	dir := t.TempDir()

	result := ReleaseResult{
		Next: "v2.0.0",
		Commits: []Commit{
			{Hash: "abc1234", Type: "feat", Subject: "remove old API", Breaking: true},
		},
	}

	if err := writeChangelog(dir, result); err != nil {
		t.Fatalf("writeChangelog: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	mustContain(t, string(data), "⚠️ BREAKING")
}

func TestRemoveChangelogEntry(t *testing.T) {
	dir := t.TempDir()
	content := "## v1.2.0 (2024-02-01)\n\n### ✨ Features\n\n- new thing\n\n## v1.1.0 (2024-01-01)\n\n### 🐛 Bug Fixes\n\n- old fix\n\n"
	os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(content), 0644)

	removed, err := removeChangelogEntry(dir, "v1.2.0")
	if err != nil {
		t.Fatalf("removeChangelogEntry: %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	result := string(data)

	if strings.Contains(result, "v1.2.0") {
		t.Error("v1.2.0 should have been removed")
	}
	mustContain(t, result, "v1.1.0")
}

func TestRemoveChangelogEntryLastOne(t *testing.T) {
	dir := t.TempDir()
	content := "## v1.0.0 (2024-01-01)\n\n### 🐛 Bug Fixes\n\n- only entry\n\n"
	os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(content), 0644)

	removed, err := removeChangelogEntry(dir, "v1.0.0")
	if err != nil {
		t.Fatalf("removeChangelogEntry: %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if strings.Contains(string(data), "v1.0.0") {
		t.Error("v1.0.0 should have been removed")
	}
}

func TestRemoveChangelogEntryNotFound(t *testing.T) {
	dir := t.TempDir()
	content := "## v1.0.0 (2024-01-01)\n\n- something\n\n"
	os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(content), 0644)

	removed, err := removeChangelogEntry(dir, "v9.9.9")
	if err != nil {
		t.Fatalf("removeChangelogEntry: %v", err)
	}
	if removed {
		t.Error("expected removed=false for non-existent entry")
	}
}

func TestRemoveChangelogEntryNoFile(t *testing.T) {
	dir := t.TempDir()

	removed, err := removeChangelogEntry(dir, "v1.0.0")
	if err != nil {
		t.Fatalf("expected no error for missing file: %v", err)
	}
	if removed {
		t.Error("expected removed=false for missing file")
	}
}

func TestFilterByType(t *testing.T) {
	commits := []Commit{
		{Type: "feat"},
		{Type: "fix"},
		{Type: "feat"},
		{Type: "unknown-type"},
	}

	feats := filterByType(commits, "feat")
	if len(feats) != 2 {
		t.Errorf("feat: got %d want 2", len(feats))
	}

	fixes := filterByType(commits, "fix")
	if len(fixes) != 1 {
		t.Errorf("fix: got %d want 1", len(fixes))
	}

	others := filterByType(commits, "other")
	if len(others) != 1 {
		t.Errorf("other: got %d want 1 (unknown-type goes to other)", len(others))
	}
}

func mustContain(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("expected to contain %q\ngot:\n%s", substr, content)
	}
}
