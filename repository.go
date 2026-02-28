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
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/Masterminds/semver/v3"
)

func openRepository(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func latestTag(repo *git.Repository) (string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "0.0.0", nil
	}

	var latest *semver.Version
	tags.ForEach(func(ref *plumbing.Reference) error {
		raw := strings.TrimPrefix(ref.Name().Short(), "v")
		v, err := semver.NewVersion(raw)
		if err != nil {
			return nil
		}
		if latest == nil || v.GreaterThan(latest) {
			latest = v
		}
		return nil
	})

	if latest == nil {
		return "0.0.0", nil
	}
	return latest.Original(), nil
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

func latestTagName(repo *git.Repository) (string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "", nil
	}

	var latest *semver.Version
	name := ""
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		raw := strings.TrimPrefix(ref.Name().Short(), "v")
		v, err := semver.NewVersion(raw)
		if err != nil {
			return nil
		}
		if latest == nil || v.GreaterThan(latest) {
			latest = v
			name = ref.Name().Short()
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return name, nil
}

func deleteTag(repo *git.Repository, tagName string) error {
	return repo.Storer.RemoveReference(plumbing.NewTagReferenceName(tagName))
}