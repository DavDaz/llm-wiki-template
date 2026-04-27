# AGENTS.md

## What This Repo Actually Does
- `llm-wiki` is a Go CLI/TUI that generates and migrates wiki directories for Claude Code, OpenCode, and Pi.
- The binary entrypoint is `cmd/llm-wiki/main.go`; behavior is wired in `internal/cmd/`.

## Commands
- `go test ./...` — full verified test suite.
- `go test ./internal/generator -run Test` — focused generator coverage.
- `go test ./internal/manifest -run Test` — focused manifest coverage.
- `make test` — full suite with `-race`.
- `make lint` — `golangci-lint run ./...`.
- `make release VERSION=v0.3.0` — tags and pushes; GitHub Actions + GoReleaser publish binaries and update the Homebrew tap.

## Architecture
- `internal/cmd/root.go`: no-arg behavior is context-sensitive.
  - inside a wiki (`wiki.toml` present) → dashboard TUI
  - outside a wiki → launcher, then guide or init wizard
- `internal/cmd/init.go`: headless mode requires BOTH `--name` and `--slug`; valid tool names are `claude-code`, `opencode`, `pi`, `all`.
- `internal/manifest/manifest.go`: `wiki.toml` is the single source of truth.
- `internal/generator/generator.go`: writes `wiki.toml`, `wiki/index.md`, `wiki/log.md`, `.gitignore`, then installs enabled tools.
- `internal/tools/*.go`: backend-specific output paths:
  - Claude → `CLAUDE.md` + `.claude/skills/*`
  - OpenCode → `AGENTS.md` + `.opencode/commands/*`
  - Pi → `AGENTS.md` + `.pi/prompts/*`
- `internal/templates/assets/` is the editable source for generated instructions and commands.

## Gotchas
- KEEP wizard form state in `*formValues` (`internal/tui/wizard/wizard.go`). Bubble Tea passes models by value; moving bound huh fields onto the model breaks pointer-based form bindings.
- `AGENTS.md` is shared by OpenCode and Pi. Uninstall logic intentionally removes it only when both are disabled.
- `generator.Migrate` is the path that re-renders instruction files after manifest changes; changing tool enablement without migrate leaves the filesystem stale.
- `manifest.Validate()` requires at least one enabled tool and a kebab-case slug.

## Read This First
- Start with `README.md`, `Makefile`, `.goreleaser.yaml`, `internal/cmd/root.go`.
- Then read only the package you are changing: `internal/cmd/`, `internal/generator/`, `internal/manifest/`, `internal/tools/`, `internal/templates/`, or `internal/tui/`.
- For generator/tooling work, use `internal/generator/generator_test.go`, `internal/manifest/manifest_test.go`, and `internal/tools/tools_test.go` as behavior specs.

## Verification
- For generator, manifest, or tool-install changes: run the focused package test plus `go test ./...`.
- For template changes: run at least `go test ./internal/generator -run Test` because those tests assert generated files and content markers.
- For command wiring changes: run `go test ./...`.

## Do Not
- Do NOT edit generated wiki output as source; change `internal/templates/assets/*` or installer logic instead.
- Do NOT change tool names or output paths casually; they are hard-coded across parsing, registry lookup, install paths, and tests.
- Do NOT treat `make release` as a dry run.

## Sources of Truth
- CLI flow: `internal/cmd/*.go`
- Manifest rules: `internal/manifest/manifest.go`
- Generation/migration: `internal/generator/generator.go`
- Tool install layout: `internal/tools/*.go`
- Generated content source: `internal/templates/assets/*`
- Release flow: `Makefile`, `.goreleaser.yaml`, `.github/workflows/release.yml`
