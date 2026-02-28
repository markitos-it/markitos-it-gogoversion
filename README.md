# gogoversion · ggv

> **The Way of the Artisan** — markitos devsecops kulture

Automatic semantic versioning from [Conventional Commits](https://www.conventionalcommits.org).
No Node. No npm. Pure Go.

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## EN — English

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
go install github.com/markitos-it/markitos-it-gogoversion@latest
```

Or clone and build:

```bash
git clone git@github.com:markitos-it/markitos-it-gogoversion.git
cd markitos-it-gogoversion
make install        # installs gogoversion + ggv symlink
```

### Usage

```bash
ggv                  # full release: bump + changelog + tag
ggv --dry-run        # preview only, no writes
ggv --no-tag         # changelog only
ggv --no-changelog   # tag only
ggv --path /my/repo  # different repo path
```

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

---

## ES — Español

### Qué hace

- Lee el historial git desde el último tag
- Parsea [Conventional Commits](https://www.conventionalcommits.org) (`feat`, `fix`, `chore`…)
- Incrementa la versión siguiendo las reglas de [SemVer](https://semver.org)
- Escribe `CHANGELOG.md`
- Crea el nuevo tag git

| Tipo de commit | Incremento |
|---|---|
| `BREAKING CHANGE` / `!` | MAJOR `x.0.0` |
| `feat` | MINOR `0.x.0` |
| `fix`, `chore`, otros | PATCH `0.0.x` |

### Instalación

```bash
go install github.com/markitos-it/markitos-it-gogoversion@latest
```

O clona y compila:

```bash
git clone git@github.com:markitos-it/markitos-it-gogoversion.git
cd markitos-it-gogoversion
make install        # instala gogoversion + symlink ggv
```

### Uso

```bash
ggv                  # release completa: bump + changelog + tag
ggv --dry-run        # previsualiza sin escribir nada
ggv --no-tag         # solo escribe el changelog
ggv --no-changelog   # solo crea el tag
ggv --path /mi/repo  # repositorio en otra ruta
```

### Conventional Commits — referencia rápida

```
feat(auth): add oauth2 login        → bump MINOR
fix: correct null pointer           → bump PATCH
feat!: remove legacy API            → bump MAJOR
fix(api)!: breaking endpoint change → bump MAJOR
```

### Desinstalar

```bash
make uninstall
```

---

## Author

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)

---

*MIT License — do whatever you want with it.*