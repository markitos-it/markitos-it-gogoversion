//#:[.'.]:>-==================================================================================
//#:[.'.]:>- Marco Antonio - markitos devsecops kulture
//#:[.'.]:>- The Way of the Artisan
//#:[.'.]:>- markitos.es.info@gmail.com
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-it/repositories
//#:[.'.]:>- 🌍 https://github.com/orgs/markitos-public/repositories
//#:[.'.]:>- 📺 https://www.youtube.com/@markitos_devsecops
//#:[.'.]:>- =================================================================================

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUndoLastReleaseNoTags(t *testing.T) {
	_, dir := initTestRepo(t)

	// No tags: should succeed and print message (nothing to undo)
	if err := undoLastRelease(dir); err != nil {
		t.Fatalf("undoLastRelease: %v", err)
	}
}

func TestUndoLastReleaseDeletesTag(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	if err := undoLastRelease(dir); err != nil {
		t.Fatalf("undoLastRelease: %v", err)
	}

	name, err := latestTagName(repo)
	if err != nil {
		t.Fatalf("latestTagName: %v", err)
	}
	if name != "" {
		t.Errorf("expected no tags after undo, got %q", name)
	}
}

func TestUndoLastReleaseRemovesChangelogEntry(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.2.0")

	content := "## v1.2.0 (2024-01-01)\n\n### ✨ Features\n\n- something\n\n## v1.0.0 (2023-01-01)\n\n- old entry\n\n"
	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := undoLastRelease(dir); err != nil {
		t.Fatalf("undoLastRelease: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	result := string(data)
	if strings.Contains(result, "v1.2.0") {
		t.Error("expected v1.2.0 to be removed from CHANGELOG.md")
	}
	if !strings.Contains(result, "v1.0.0") {
		t.Error("expected v1.0.0 to remain in CHANGELOG.md")
	}
}

func TestUndoLastReleasePicksHighestTag(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")
	addTag(t, repo, "v2.0.0")

	if err := undoLastRelease(dir); err != nil {
		t.Fatalf("undoLastRelease: %v", err)
	}

	// After undo, v2.0.0 should be gone, v1.0.0 should remain
	name, err := latestTagName(repo)
	if err != nil {
		t.Fatalf("latestTagName: %v", err)
	}
	if name != "v1.0.0" {
		t.Errorf("got %q want %q", name, "v1.0.0")
	}
}

func TestUndoLastReleaseInvalidPath(t *testing.T) {
	dir := t.TempDir() // not a git repo
	err := undoLastRelease(dir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}
