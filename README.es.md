# gogoversion · ggv

> **The Way of the Artisan** — markitos devsecops kulture

Versionado semántico automático desde [Conventional Commits](https://www.conventionalcommits.org).
Sin Node. Sin npm. Go puro.

[![Go](https://img.shields.io/badge/go-1.25-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

Versión en inglés: [README.md](README.md) · Para contribuidores: [DEVELOPERS.md](DEVELOPERS.md)

---

## Instalación

```bash
go install github.com/markitos-it/markitos-it-gogoversion/cmd/gogoversion@latest
```

El binario quedará disponible como `gogoversion` en tu `$GOPATH/bin`.

> Asegúrate de tener `$(go env GOPATH)/bin` en tu `$PATH`.

### Alias opcional `ggv`

```bash
ln -sf "$(go env GOPATH)/bin/gogoversion" "$(go env GOPATH)/bin/ggv"
```

---

## Qué hace

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

---

## Uso

```bash
gogoversion .                    # release completa en repo actual
gogoversion --dry-run .          # previsualiza sin escribir nada
gogoversion --no-tag .           # solo escribe el changelog
gogoversion --no-changelog .     # solo crea el tag
gogoversion --undo .             # deshace el último release (tag + changelog)
gogoversion --dry-run /mi/repo   # repositorio en otra ruta
gogoversion -h | --help          # muestra ayuda
gogoversion --version            # muestra la versión del binario
```

Con el alias `ggv`:

```bash
ggv .
ggv --dry-run .
```

`repo_path` es obligatorio, debe ir siempre al final y todas las opciones van antes.

En modo interactivo (`gogoversion .`), la herramienta pide el tipo y mensaje del commit de release, ejecuta `git add CHANGELOG.md` + `git commit`, crea el tag y después hace push de la rama actual y del tag a `origin`.

---

## Conventional Commits — referencia rápida

```
feat(auth): add oauth2 login        → bump MINOR
fix: correct null pointer           → bump PATCH
feat!: remove legacy API            → bump MAJOR
fix(api)!: breaking endpoint change → bump MAJOR
```

---

## Autor

**Marco Antonio** — [markitos devsecops kulture](https://github.com/orgs/markitos-it/repositories)
📺 [youtube.com/@markitos_devsecops](https://www.youtube.com/@markitos_devsecops)

---

*MIT License — do whatever you want with it.*
