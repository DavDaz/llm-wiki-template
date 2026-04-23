# llm-wiki — Contexto del proyecto

## Qué es esto

Generador de wikis de conocimiento mantenidos por IA. Un binario Go (`llm-wiki`) reemplaza el viejo `setup.sh`. Distribuido via Homebrew.

## Cómo publicar una nueva release

```bash
git tag v0.X.0
git push origin v0.X.0
```

GitHub Actions corre GoReleaser automáticamente:
- Compila binarios para darwin/linux (amd64 + arm64) y windows
- Publica el release en GitHub con los assets
- Actualiza la fórmula Homebrew en `DavDaz/homebrew-llm-wiki`

## Cómo instalar (usuario final)

```bash
brew tap DavDaz/llm-wiki
brew install llm-wiki
```

O con Go:

```bash
go install github.com/DavDaz/llm-wiki-generator/cmd/llm-wiki@latest
```

## Repositorios involucrados

| Repo | Propósito |
|------|-----------|
| `DavDaz/llm-wiki-generator` | Repo principal — código fuente |
| `DavDaz/homebrew-llm-wiki` | Tap de Homebrew — fórmula auto-generada por GoReleaser |

## Secrets requeridos en llm-wiki-generator

| Secret | Para qué |
|--------|----------|
| `HOMEBREW_TAP_TOKEN` | Fine-grained PAT con Contents R/W sobre `homebrew-llm-wiki` |

## Estructura del proyecto

```
cmd/llm-wiki/          → entrypoint del binario
internal/
  cmd/                 → comandos Cobra (init, manage, status, add-tool, remove-tool, migrate)
  generator/           → crea y migra wikis en el filesystem
  manifest/            → lee/escribe wiki.toml
  templates/           → archivos embebidos (CLAUDE.md, AGENTS.md, commands/)
  tools/               → registry de tool backends (claude-code, opencode, pi)
  tui/
    wizard/            → form TUI para llm-wiki init (huh + bubbletea)
    dashboard/         → panel de gestión para llm-wiki manage (bubbletea)
    styles/            → estilos Lipgloss compartidos
  version/             → versión inyectada por GoReleaser via ldflags
assets/                → GUIDE.md y templates fuente (referencia)
.goreleaser.yaml       → config de build y distribución
.github/workflows/     → release.yml dispara GoReleaser en push de tags v*
```

## Comandos disponibles

```bash
llm-wiki init                          # wizard TUI para crear wiki nuevo
llm-wiki init --name X --slug x        # modo headless
llm-wiki manage                        # dashboard TUI para gestionar wiki
llm-wiki status                        # estado del wiki actual
llm-wiki add-tool opencode             # habilitar tool backend
llm-wiki remove-tool pi                # deshabilitar tool backend
llm-wiki migrate                       # sincronizar manifest con filesystem
llm-wiki version                       # imprimir versión
```

## Nota técnica importante — bubbletea + huh

Los valores del form TUI viven en un `*formValues` (puntero al heap), no como campos del `Model`. Esto es necesario porque Bubbletea pasa el Model por valor y los punteros que huh necesita para hacer binding quedan inválidos si están en el stack. No cambiar esto sin entender la razón.
