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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
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

func TestUndoLastReleaseDeletesRemoteTag(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	remoteDir := t.TempDir()
	if _, err := git.PlainInit(remoteDir, true); err != nil {
		t.Fatalf("PlainInit bare remote: %v", err)
	}
	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remoteDir}}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	branch := head.Name().Short()
	branchSpec := gitconfig.RefSpec("refs/heads/" + branch + ":refs/heads/" + branch)
	tagSpec := gitconfig.RefSpec("refs/tags/v1.0.0:refs/tags/v1.0.0")
	if err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []gitconfig.RefSpec{branchSpec, tagSpec}}); err != nil {
		t.Fatalf("Push: %v", err)
	}

	if err := undoLastRelease(dir); err != nil {
		t.Fatalf("undoLastRelease: %v", err)
	}

	remoteRepo, err := git.PlainOpen(remoteDir)
	if err != nil {
		t.Fatalf("PlainOpen remote: %v", err)
	}
	if _, err := remoteRepo.Reference(plumbing.NewTagReferenceName("v1.0.0"), true); err == nil {
		t.Fatal("expected remote tag v1.0.0 to be deleted")
	}
}

func TestUndoLastReleaseLatestTagNameError(t *testing.T) {
	_, dir := initTestRepo(t)

	boom := errors.New("latestTagName boom")
	orig := undoLatestTagNameFn
	undoLatestTagNameFn = func(_ *git.Repository) (string, error) {
		return "", boom
	}
	t.Cleanup(func() { undoLatestTagNameFn = orig })

	err := undoLastRelease(dir)
	if err == nil {
		t.Fatal("expected error from latestTagName")
	}
}

func TestUndoLastReleaseDeleteTagError(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	boom := errors.New("deleteTag boom")
	orig := undoDeleteTagFn
	undoDeleteTagFn = func(_ *git.Repository, _ string) error { return boom }
	t.Cleanup(func() { undoDeleteTagFn = orig })

	err := undoLastRelease(dir)
	if err == nil || err.Error() != boom.Error() {
		t.Fatalf("expected deleteTag error, got %v", err)
	}
}

func TestUndoLastReleaseChangelogReadError(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	// make CHANGELOG.md a directory so os.ReadFile returns an error
	clPath := filepath.Join(dir, "CHANGELOG.md")
	if err := os.Mkdir(clPath, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	err := undoLastRelease(dir)
	if err == nil {
		t.Fatal("expected error when CHANGELOG.md is a directory")
	}
}
