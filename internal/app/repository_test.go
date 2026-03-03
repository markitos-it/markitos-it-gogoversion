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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// initTestRepo creates a non-bare git repository in a temp dir with an initial commit.
func initTestRepo(t *testing.T) (*git.Repository, string) {
	t.Helper()
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	commitFile(t, repo, dir, "README.md", "init")
	return repo, dir
}

// commitFile writes content to filename inside dir and creates a commit.
func commitFile(t *testing.T, repo *git.Repository, dir, filename, msg string) plumbing.Hash {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(msg), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if _, err := wt.Add(filename); err != nil {
		t.Fatalf("Add: %v", err)
	}
	sig := &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()}
	hash, err := wt.Commit(msg, &git.CommitOptions{Author: sig, Committer: sig})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return hash
}

// addTag creates a lightweight tag on HEAD of the given repo.
func addTag(t *testing.T, repo *git.Repository, tagName string) {
	t.Helper()
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	_, err = repo.CreateTag(tagName, head.Hash(), nil)
	if err != nil {
		t.Fatalf("CreateTag %q: %v", tagName, err)
	}
}

func TestOpenRepository(t *testing.T) {
	_, dir := initTestRepo(t)

	repo, err := openRepository(dir)
	if err != nil {
		t.Fatalf("openRepository: %v", err)
	}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestOpenRepositoryInvalid(t *testing.T) {
	dir := t.TempDir()
	_, err := openRepository(dir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestLatestTagNoTags(t *testing.T) {
	repo, _ := initTestRepo(t)

	tag, err := latestTag(repo)
	if err != nil {
		t.Fatalf("latestTag: %v", err)
	}
	if tag != "0.0.0" {
		t.Errorf("got %q want %q", tag, "0.0.0")
	}
}

func TestLatestTagWithTags(t *testing.T) {
	repo, _ := initTestRepo(t)
	addTag(t, repo, "v1.2.0")
	addTag(t, repo, "v1.0.0")

	tag, err := latestTag(repo)
	if err != nil {
		t.Fatalf("latestTag: %v", err)
	}
	if tag != "1.2.0" {
		t.Errorf("got %q want %q", tag, "1.2.0")
	}
}

func TestLatestTagNameWithTags(t *testing.T) {
	repo, _ := initTestRepo(t)
	addTag(t, repo, "v2.1.0")

	name, err := latestTagName(repo)
	if err != nil {
		t.Fatalf("latestTagName: %v", err)
	}
	if name != "v2.1.0" {
		t.Errorf("got %q want %q", name, "v2.1.0")
	}
}

func TestCreateTag(t *testing.T) {
	repo, _ := initTestRepo(t)

	if err := createTag(repo, "v1.0.0"); err != nil {
		t.Fatalf("createTag: %v", err)
	}

	name, err := latestTagName(repo)
	if err != nil {
		t.Fatalf("latestTagName: %v", err)
	}
	if name != "v1.0.0" {
		t.Errorf("got %q want %q", name, "v1.0.0")
	}
}

func TestDeleteTag(t *testing.T) {
	repo, _ := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	if err := deleteTag(repo, "v1.0.0"); err != nil {
		t.Fatalf("deleteTag: %v", err)
	}

	found := false
	tags, _ := repo.Tags()
	tags.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == "v1.0.0" {
			found = true
		}
		return nil
	})
	if found {
		t.Error("expected tag v1.0.0 to be deleted")
	}
}

func TestCommitsSinceTag(t *testing.T) {
	repo, dir := initTestRepo(t)
	addTag(t, repo, "v1.0.0")

	commitFile(t, repo, dir, "a.txt", "feat: add a")
	commitFile(t, repo, dir, "b.txt", "fix: fix b")

	commits, err := commitsSinceTag(repo, "1.0.0")
	if err != nil {
		t.Fatalf("commitsSinceTag: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("got %d commits want 2", len(commits))
	}
}

func TestCommitsSinceTagNoTag(t *testing.T) {
	repo, dir := initTestRepo(t)
	commitFile(t, repo, dir, "a.txt", "feat: add a")

	commits, err := commitsSinceTag(repo, "0.0.0")
	if err != nil {
		t.Fatalf("commitsSinceTag: %v", err)
	}
	// all commits since no matching tag
	if len(commits) == 0 {
		t.Error("expected at least one commit")
	}
}

func TestAddAndCommitChangelog(t *testing.T) {
	repo, dir := initTestRepo(t)

	origAuthorName := os.Getenv("GIT_AUTHOR_NAME")
	origAuthorEmail := os.Getenv("GIT_AUTHOR_EMAIL")
	origCommitterName := os.Getenv("GIT_COMMITTER_NAME")
	origCommitterEmail := os.Getenv("GIT_COMMITTER_EMAIL")
	os.Setenv("GIT_AUTHOR_NAME", "Test User")
	os.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	os.Unsetenv("GIT_COMMITTER_NAME")
	os.Unsetenv("GIT_COMMITTER_EMAIL")
	defer func() {
		os.Setenv("GIT_AUTHOR_NAME", origAuthorName)
		os.Setenv("GIT_AUTHOR_EMAIL", origAuthorEmail)
		os.Setenv("GIT_COMMITTER_NAME", origCommitterName)
		os.Setenv("GIT_COMMITTER_EMAIL", origCommitterEmail)
	}()

	if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte("# Changelog\n"), 0644); err != nil {
		t.Fatalf("WriteFile CHANGELOG.md: %v", err)
	}

	if err := addAndCommitChangelog(repo, "chore(release): prepara release v1.0.0"); err != nil {
		t.Fatalf("addAndCommitChangelog: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	commitObj, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	if commitObj.Message != "chore(release): prepara release v1.0.0" {
		t.Errorf("got commit message %q", commitObj.Message)
	}
}

func TestResolveGitSignatureWithoutIdentity(t *testing.T) {
	repo, _ := initTestRepo(t)

	origAuthorName := os.Getenv("GIT_AUTHOR_NAME")
	origAuthorEmail := os.Getenv("GIT_AUTHOR_EMAIL")
	origCommitterName := os.Getenv("GIT_COMMITTER_NAME")
	origCommitterEmail := os.Getenv("GIT_COMMITTER_EMAIL")
	os.Unsetenv("GIT_AUTHOR_NAME")
	os.Unsetenv("GIT_AUTHOR_EMAIL")
	os.Unsetenv("GIT_COMMITTER_NAME")
	os.Unsetenv("GIT_COMMITTER_EMAIL")
	defer func() {
		os.Setenv("GIT_AUTHOR_NAME", origAuthorName)
		os.Setenv("GIT_AUTHOR_EMAIL", origAuthorEmail)
		os.Setenv("GIT_COMMITTER_NAME", origCommitterName)
		os.Setenv("GIT_COMMITTER_EMAIL", origCommitterEmail)
	}()

	_, err := resolveGitSignature(repo)
	if err == nil {
		t.Fatal("expected error when git identity is missing")
	}
}

func TestAddFilesAndCommitChanges(t *testing.T) {
	repo, dir := initTestRepo(t)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatalf("WriteFile a.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatalf("WriteFile b.txt: %v", err)
	}

	if err := addFiles(repo, []string{"a.txt", "b.txt"}); err != nil {
		t.Fatalf("addFiles: %v", err)
	}

	message := "chore: add files"
	if err := commitChanges(repo, message); err != nil {
		t.Fatalf("commitChanges: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	commitObj, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	if commitObj.Message != message {
		t.Fatalf("got %q want %q", commitObj.Message, message)
	}
}

func TestAddAllChangedFiles(t *testing.T) {
	repo, dir := initTestRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	count, err := addAllChangedFiles(repo)
	if err != nil {
		t.Fatalf("addAllChangedFiles: %v", err)
	}
	if count != 1 {
		t.Fatalf("got %d want %d", count, 1)
	}
}

func TestPullCurrentBranchDetachedHead(t *testing.T) {
	repo, dir := initTestRepo(t)
	hash := commitFile(t, repo, dir, "detached.txt", "detached")

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := wt.Checkout(&git.CheckoutOptions{Hash: hash}); err != nil {
		t.Fatalf("Checkout: %v", err)
	}

	_, err = pullCurrentBranch(repo)
	if err == nil {
		t.Fatal("expected error on detached HEAD")
	}
}

func TestPullCurrentBranchWithoutOrigin(t *testing.T) {
	repo, _ := initTestRepo(t)
	_, err := pullCurrentBranch(repo)
	if err == nil {
		t.Fatal("expected error when origin is not configured")
	}
}

func TestPushReleaseWithAndWithoutTag(t *testing.T) {
	remoteDir := t.TempDir()
	if _, err := git.PlainInit(remoteDir, true); err != nil {
		t.Fatalf("PlainInit bare remote: %v", err)
	}

	repo, dir := initTestRepo(t)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remoteDir}}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "push.txt"), []byte("push"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := addFiles(repo, []string{"push.txt"}); err != nil {
		t.Fatalf("addFiles: %v", err)
	}
	if err := commitChanges(repo, "chore: push test"); err != nil {
		t.Fatalf("commitChanges: %v", err)
	}
	if err := createTag(repo, "v9.9.9"); err != nil {
		t.Fatalf("createTag: %v", err)
	}

	if err := pushRelease(repo, "v9.9.9", false); err != nil {
		t.Fatalf("pushRelease includeTag=false: %v", err)
	}

	remoteRepo, err := git.PlainOpen(remoteDir)
	if err != nil {
		t.Fatalf("PlainOpen remote: %v", err)
	}

	branch, err := currentBranchName(repo)
	if err != nil {
		t.Fatalf("currentBranchName: %v", err)
	}
	if _, err := remoteRepo.Reference(plumbing.NewBranchReferenceName(branch), true); err != nil {
		t.Fatalf("expected branch to exist on remote: %v", err)
	}

	if _, err := remoteRepo.Reference(plumbing.NewTagReferenceName("v9.9.9"), true); err == nil {
		t.Fatal("expected tag not to be pushed when includeTag=false")
	}

	if err := pushRelease(repo, "v9.9.9", true); err != nil {
		t.Fatalf("pushRelease includeTag=true: %v", err)
	}
	if _, err := remoteRepo.Reference(plumbing.NewTagReferenceName("v9.9.9"), true); err != nil {
		t.Fatalf("expected tag to be pushed: %v", err)
	}
}

func TestPushReleaseDetachedHead(t *testing.T) {
	remoteDir := t.TempDir()
	if _, err := git.PlainInit(remoteDir, true); err != nil {
		t.Fatalf("PlainInit bare remote: %v", err)
	}

	repo, dir := initTestRepo(t)
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remoteDir}}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}

	hash := commitFile(t, repo, dir, "detached-push.txt", fmt.Sprintf("detached-%d", time.Now().UnixNano()))
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := wt.Checkout(&git.CheckoutOptions{Hash: hash}); err != nil {
		t.Fatalf("Checkout detached: %v", err)
	}

	err = pushRelease(repo, "v0.0.1", false)
	if err == nil {
		t.Fatal("expected error on detached HEAD")
	}
}
