# gogoversion · ggv

> **The Way of the Artisan** — markitos devsecops kulture

Versión en inglés: [README.md](README.md)

Versionado semántico automático desde [Conventional Commits](https://www.conventionalcommits.org).
Sin Node. Sin npm. Go puro.

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

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
go build -o gogoversion ./cmd/gogoversion
cp gogoversion "$(go env GOPATH)/bin/gogoversion"
ln -sf "$(go env GOPATH)/bin/gogoversion" "$(go env GOPATH)/bin/ggv"
```

O desde este repo:

```bash
make install        # instala gogoversion + symlink ggv
```

O clona y compila:

```bash
git clone git@github.com:markitos-it/markitos-it-gogoversion.git
cd markitos-it-gogoversion
make install        # instala gogoversion + symlink ggv
```

### Uso

```bash
ggv .                            # release completa en repo actual
ggv --dry-run .                  # previsualiza sin escribir nada
ggv --no-tag .                   # solo escribe el changelog
ggv --no-changelog .             # solo crea el tag
ggv --undo .                     # deshace el último release (tag + changelog)
ggv --dry-run /mi/repo           # repositorio en otra ruta
ggv -h | ggv --help              # muestra ayuda
ggv --version                    # muestra la versión del binario
```

`repo_path` es obligatorio, debe ir siempre al final y todas las opciones van antes.

En modo interactivo (`ggv .`), la herramienta pide el tipo y mensaje del commit de release, ejecuta `git add CHANGELOG.md` + `git commit`, crea el tag y después hace push de la rama actual y del tag a `origin`.

El modo release permite cambios locales pendientes (flujo normal durante desarrollo).

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

### Targets de Make

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

## Autor

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)

---

*MIT License — do whatever you want with it.*
