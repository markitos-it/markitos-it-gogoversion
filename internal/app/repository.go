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
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type tagInfo struct {
	Name    string
	Version *semver.Version
}

func openRepository(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func resolveLatestTag(repo *git.Repository) (tagInfo, error) {
	tags, err := repo.Tags()
	if err != nil {
		return tagInfo{}, nil
	}

	var best tagInfo
	tags.ForEach(func(ref *plumbing.Reference) error {
		raw := strings.TrimPrefix(ref.Name().Short(), "v")
		v, err := semver.NewVersion(raw)
		if err != nil {
			return nil
		}
		if best.Version == nil || v.GreaterThan(best.Version) {
			best = tagInfo{Name: ref.Name().Short(), Version: v}
		}
		return nil
	})

	return best, nil
}

func latestTag(repo *git.Repository) (string, error) {
	info, err := resolveLatestTag(repo)
	if err != nil || info.Version == nil {
		return "0.0.0", nil
	}
	return info.Version.Original(), nil
}

func latestTagName(repo *git.Repository) (string, error) {
	info, err := resolveLatestTag(repo)
	if err != nil {
		return "", err
	}
	return info.Name, nil
}

func commitsSinceTag(repo *git.Repository, tag string) ([]*object.Commit, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	tagHash := resolveTagHash(repo, tag)

	iter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	iter.ForEach(func(c *object.Commit) error {
		if c.Hash == tagHash {
			return storer.ErrStop
		}
		commits = append(commits, c)
		return nil
	})

	return commits, nil
}

func resolveTagHash(repo *git.Repository, tag string) plumbing.Hash {
	var hash plumbing.Hash
	tags, _ := repo.Tags()
	tags.ForEach(func(ref *plumbing.Reference) error {
		clean := strings.TrimPrefix(ref.Name().Short(), "v")
		if clean != tag {
			return nil
		}
		if obj, err := repo.TagObject(ref.Hash()); err == nil {
			hash = obj.Target
		} else {
			hash = ref.Hash()
		}
		return nil
	})
	return hash
}

func createTag(repo *git.Repository, version string) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}
	_, err = repo.CreateTag(version, head.Hash(), nil)
	return err
}

func deleteTag(repo *git.Repository, tagName string) error {
	return repo.Storer.RemoveReference(plumbing.NewTagReferenceName(tagName))
}

func addAndCommitChangelog(repo *git.Repository, message string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := wt.Add("CHANGELOG.md"); err != nil {
		return err
	}

	author, err := resolveGitSignature(repo)
	if err != nil {
		return err
	}

	_, err = wt.Commit(message, &git.CommitOptions{
		Author:    author,
		Committer: author,
	})
	return err
}

func addFiles(repo *git.Repository, files []string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	for _, file := range files {
		if _, err := wt.Add(file); err != nil {
			return err
		}
	}
	return nil
}

func commitChanges(repo *git.Repository, message string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	author, err := resolveGitSignature(repo)
	if err != nil {
		return err
	}

	_, err = wt.Commit(message, &git.CommitOptions{
		Author:    author,
		Committer: author,
	})
	return err
}

func addAllChangedFiles(repo *git.Repository) (int, error) {
	files, err := changedFiles(repo)
	if err != nil {
		return 0, err
	}
	if len(files) == 0 {
		return 0, nil
	}

	if err := addFiles(repo, files); err != nil {
		return 0, err
	}
	return len(files), nil
}

func resolveGitSignature(repo *git.Repository) (*object.Signature, error) {
	config, err := repo.Config()
	if err == nil && config.User.Name != "" && config.User.Email != "" {
		return &object.Signature{Name: config.User.Name, Email: config.User.Email, When: time.Now()}, nil
	}

	name := strings.TrimSpace(os.Getenv("GIT_AUTHOR_NAME"))
	email := strings.TrimSpace(os.Getenv("GIT_AUTHOR_EMAIL"))
	if name == "" {
		name = strings.TrimSpace(os.Getenv("GIT_COMMITTER_NAME"))
	}
	if email == "" {
		email = strings.TrimSpace(os.Getenv("GIT_COMMITTER_EMAIL"))
	}

	if name == "" || email == "" {
		return nil, fmt.Errorf("could not resolve git user (configure user.name and user.email)")
	}

	return &object.Signature{Name: name, Email: email, When: time.Now()}, nil
}

func currentBranchName(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "detached", nil
}

func changedFiles(repo *git.Repository) ([]string, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(status))
	for file, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified || fileStatus.Worktree != git.Unmodified {
			files = append(files, file)
		}
	}

	slices.Sort(files)
	return files, nil
}

func pullCurrentBranch(repo *git.Repository) (bool, error) {
	head, err := repo.Head()
	if err != nil {
		return false, err
	}
	if !head.Name().IsBranch() {
		return false, fmt.Errorf("HEAD is not on a branch; cannot pull automatically")
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	err = wt.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: head.Name(),
		SingleBranch:  true,
	})
	if err == git.NoErrAlreadyUpToDate {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func pushRelease(repo *git.Repository, tag string, includeTag bool) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}

	if !head.Name().IsBranch() {
		return fmt.Errorf("HEAD is not on a branch; cannot push automatically")
	}

	branch := head.Name().Short()
	branchSpec := gitconfig.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	if err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []gitconfig.RefSpec{branchSpec}}); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	if includeTag {
		tagSpec := gitconfig.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
		if err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []gitconfig.RefSpec{tagSpec}}); err != nil && err != git.NoErrAlreadyUpToDate {
			return err
		}
	}

	return nil
}

func deleteRemoteTag(repo *git.Repository, tagName string) error {
	tagSpec := gitconfig.RefSpec(fmt.Sprintf(":refs/tags/%s", tagName))
	err := repo.Push(&git.PushOptions{RemoteName: "origin", RefSpecs: []gitconfig.RefSpec{tagSpec}})
	if err == nil || err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}
