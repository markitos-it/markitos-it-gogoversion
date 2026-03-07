# gogoversion · ggv

> **The Way of the Artisan** — markitos devsecops kulture

Automatic semantic versioning from [Conventional Commits](https://www.conventionalcommits.org).
No Node. No npm. Pure Go.

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

Spanish version: [README.es.md](README.es.md) · For contributors: [DEVELOPERS.md](DEVELOPERS.md)

---

## Install

```bash
go install github.com/markitos-it/markitos-it-gogoversion/cmd/gogoversion@latest
```

The binary will be available as `gogoversion` in your `$GOPATH/bin`.

> Make sure `$(go env GOPATH)/bin` is in your `$PATH`.

### Create aliases `ggv` and `gogov`

```bash
gogoversion install
```

This creates `ggv` and `gogov` symlinks pointing to `gogoversion` in the same bin directory.

---

## What it does

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

---

## Usage

```bash
gogoversion .                    # full release on current repo
gogoversion --dry-run .          # preview only, no writes
gogoversion --no-tag .           # changelog only
gogoversion --no-changelog .     # tag only
gogoversion --undo .             # undo last release (tag + changelog entry)
gogoversion --dry-run /my/repo   # different repo path
gogoversion -h | --help          # show help
gogoversion --version            # show binary version
```

With the `ggv` alias:

```bash
ggv .
ggv --dry-run .
```

`repo_path` is required, must be the last argument, and all options must go before it.

In interactive mode (`gogoversion .`), the tool asks for the release commit type and message, runs `git add CHANGELOG.md` + `git commit`, creates the tag, and then pushes the current branch and tag to `origin`.

---

## Conventional Commits — quick reference

```
feat(auth): add oauth2 login        → MINOR bump
fix: correct null pointer           → PATCH bump
feat!: remove legacy API            → MAJOR bump
fix(api)!: breaking endpoint change → MAJOR bump
```

---

## Author

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)

---

*MIT License — do whatever you want with it.*
