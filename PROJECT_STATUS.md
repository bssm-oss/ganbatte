# ganbatte Project Status

> Last updated: 2026-04-13

## Current State: Feature-Complete Pre-Release

모든 계획된 기능이 구현 완료. CI green (3-OS matrix + lint). v1.0 태그 준비 가능 상태.

## Implemented Features

### Core (v0.1)
- [x] `gnb init` — config initialization with format selection
- [x] `gnb add` / `gnb edit` / `gnb remove` — alias CRUD
- [x] `gnb run` — execute alias or workflow
- [x] `gnb list` — list with `--tag` filtering
- [x] Multi-format config (TOML/YAML/JSON) via Viper

### v0.2
- [x] Shell detection (zsh/bash/fish)
- [x] History mining — `gnb suggest` (frequency + sequence analysis)
- [x] TUI browser — fuzzy search, tag filter, preview, confirm-delete
- [x] `gnb run --dry-run` with destructive command detection
- [x] `gnb show` — detailed alias/workflow view
- [x] `gnb doctor` — environment diagnostics
- [x] Workflow engine: `on_fail` (stop/continue/prompt), `confirm` gates

### v0.3
- [x] `gnb config path` / `gnb config convert --to <format>`
- [x] `gnb export` / `gnb import` — config portability
- [x] Dual scope: global + project with auto-merge
- [x] Conflict detection between scopes

### v1.0
- [x] Schema version management + migration
- [x] Man pages for all commands (auto-generated)
- [x] `gnb shell-init` — `eval "$(gnb shell-init)"` for native shell integration
- [x] `gnb init --project` — project-scoped config for team onboarding
- [x] `gnb migrate` — one-command shell alias import from .zshrc/.bashrc/config.fish
- [x] Parameterized aliases (`params`, `default_params`)
- [x] Confirm guard for aliases (`confirm = true`, `--yes/-y` skip)
- [x] TUI scoped display (`[project]` labels)
- [x] `gnb suggest` sectioned output + workflow apply

## Architecture

```
cmd/                    Cobra command definitions
  add.go, edit.go, remove.go, run.go, list.go, show.go,
  init.go, doctor.go, suggest.go, migrate.go,
  shell_init.go, export.go, import.go, config_*.go, root.go
internal/
  config/               Multi-format config, scoping, merge, portability, schema
  history/              Shell history parsers (zsh/bash/fish), suggest engine
  shell/                Shell detection, history paths, alias migration parser
  tui/                  Bubbletea model, item list, fuzzy filter
  workflow/             Step execution, parameter substitution, confirm/on_fail
docs/
  man/                  Auto-generated man pages (17 pages)
  spec.md               Original feature specification
testdata/fixtures/      Shell history samples, config format examples
```

## CI/CD

- GitHub Actions: test (ubuntu/macos/windows) + lint (golangci-lint)
- goreleaser: multi-OS binaries + Homebrew formula tap (`bssm-oss/homebrew-tap`)
- All tests use `shell: bash` on Windows to avoid PowerShell parsing issues

## Test Coverage

- 87+ test cases across all packages
- Race detector enabled (`-race`)
- Fixture-based format equivalence tests
- Mock executor for workflow unit tests

## Dependencies

- `spf13/cobra` v1.10.2 — CLI framework
- `spf13/viper` v1.18.0 — config management
- `charmbracelet/bubbletea` v1.3.10 — TUI
- `charmbracelet/bubbles` v1.0.0 — TUI components
- `charmbracelet/lipgloss` v1.1.0 — styling
- `sahilm/fuzzy` v0.1.1 — fuzzy search
- `stretchr/testify` v1.8.4 — test assertions
