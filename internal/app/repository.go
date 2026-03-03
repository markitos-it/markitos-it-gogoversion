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
		return nil, fmt.Errorf("no se pudo resolver usuario git (configura user.name y user.email)")
	}

	return &object.Signature{Name: name, Email: email, When: time.Now()}, nil
}

func ensureCleanWorktree(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := wt.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		return nil
	}

	files := make([]string, 0, len(status))
	for file := range status {
		files = append(files, file)
	}

	return fmt.Errorf("working tree con cambios pendientes: %s", strings.Join(files, ", "))
}

func pushRelease(repo *git.Repository, tag string, includeTag bool) error {
	head, err := repo.Head()
	if err != nil {
		return err
	}

	if !head.Name().IsBranch() {
		return fmt.Errorf("HEAD no está en una rama; no se puede hacer push automático")
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
