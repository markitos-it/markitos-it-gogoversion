# DEVELOPERS.md — gogoversion

> Guide for contributors and developers working on `gogoversion`.

---

## Requirements

- Go 1.25+
- Git

---

## Clone and build

```bash
git clone git@github.com:markitos-it/markitos-it-gogoversion.git
cd markitos-it-gogoversion
make build
```

---

## Install locally (from source)

```bash
make install
```

This builds the binary and copies it to `$(go env GOPATH)/bin/gogoversion`, plus creates a `ggv` symlink.

## Uninstall

```bash
make uninstall
```

---

## Make targets

```bash
make help           # list all available targets
make build          # build binary
make tidy           # go mod tidy
make install        # build + install binary + ggv symlink
make uninstall      # remove installed binaries and symlink
make run-dry        # run --dry-run on current repo
make run-real       # run full release on current repo
make run-no-tag     # run --no-tag on current repo
make run-no-changelog  # run --no-changelog on current repo
make test           # run tests
make test-v         # run tests with verbose output
make cover          # run tests with coverage report
make clean          # remove local binary artifacts
make clean-cache    # clean Go build cache
```

---

## Project structure

```
cmd/gogoversion/    # main entry point
internal/app/       # core application logic
```

---

## Running tests

```bash
make test
make test-v
make cover
```

---

## Release process

Releases are automated via GitHub Actions using `gogoversion` itself.
To create a new release, run `gogoversion` (or `ggv`) on the repo and follow the interactive prompts.

---

## Author

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)
