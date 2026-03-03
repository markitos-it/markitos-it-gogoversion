# gogoversion · ggv

> **The Way of the Artisan** — markitos devsecops kulture

Automatic semantic versioning from [Conventional Commits](https://www.conventionalcommits.org).
No Node. No npm. Pure Go.

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## EN — English

Spanish version: [README.es.md](README.es.md)

### What it does

- Reads the git history since the last tag
- Parses [Conventional Commits](https://www.conventionalcommits.org) (`feat`, `fix`, `chore`…)
- Bumps the version following [SemVer](https://semver.org) rules
- Writes `CHANGELOG.md`
- Creates the new git tag

| Commit type | Version bump |
|---|---|
| `BREAKING CHANGE` / `!` | MAJOR `x.0.0` |
| `feat` | MINOR `0.x.0` |
| `fix`, `chore`, others | PATCH `0.0.x` |

### Install

```bash
go build -o gogoversion ./cmd/app
cp gogoversion "$(go env GOPATH)/bin/gogoversion"
ln -sf "$(go env GOPATH)/bin/gogoversion" "$(go env GOPATH)/bin/ggv"
```

Or from this repo:

```bash
make install        # installs gogoversion + ggv symlink
```

Or clone and build:

```bash
git clone git@github.com:markitos-it/markitos-it-gogoversion.git
cd markitos-it-gogoversion
make install        # installs gogoversion + ggv symlink
```

### Usage

```bash
ggv .                            # full release on current repo
ggv --dry-run .                  # preview only, no writes
ggv --no-tag .                   # changelog only
ggv --no-changelog .             # tag only
ggv --undo .                     # undo last release (tag + changelog entry)
ggv --dry-run /my/repo           # different repo path
ggv -h | ggv --help  # show help
ggv --version        # show binary version
```

`repo_path` is required, must be the last argument, and all options must go before it.

In interactive mode (`ggv .`), the tool asks for the release commit type and message, runs `git add CHANGELOG.md` + `git commit`, creates the tag, and then pushes the current branch and tag to `origin`.

Release mode allows local pending changes (normal workflow while developing).

### Conventional Commits — quick reference

```
feat(auth): add oauth2 login        → MINOR bump
fix: correct null pointer           → PATCH bump
feat!: remove legacy API            → MAJOR bump
fix(api)!: breaking endpoint change → MAJOR bump
```

### Uninstall

```bash
make uninstall
```

### Make targets

```bash
make help
make build
make tidy
make run-dry
make run-real
make run-no-tag
make run-no-changelog
make test
make test-v
make cover
make clean
make clean-cache
```

---

## Author

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)

---

*MIT License — do whatever you want with it.*