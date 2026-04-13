# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- `gnb migrate` — import existing shell aliases from .zshrc/.bashrc/config.fish in one command
- `gnb shell-init` — output shell functions for `eval "$(gnb shell-init)"` integration (bash/zsh/fish)
- `gnb init --project` — create project-scoped `.ganbatte.*` config for team onboarding
- Parameterized aliases — `params` and `default_params` fields on Alias struct
- Confirm guard for aliases — `confirm = true` prompts before execution
- `gnb run --yes/-y` flag to skip confirmation prompts (CI-friendly)
- TUI: scoped display with `[project]` labels, project items shown first
- `gnb suggest`: sectioned output (alias vs workflow), `--apply` now creates workflows too
- `on_fail: "prompt"` — interactive continue/abort prompt on workflow step failure

### Fixed
- Cross-platform: `portability.go` uses `os.CreateTemp` instead of `/tmp` hardcode
- `SaveGlobal()` now detects existing config format (yaml/json) instead of always writing TOML
- Schema version: `gnb init` templates use `"1.0.0"` matching `SchemaVersion` constant
- Windows CI: set `USERPROFILE` alongside `HOME` in all tests
- CI: `shell: bash` in test matrix to fix PowerShell `-coverprofile` parsing
- All golangci-lint warnings resolved (octal literals, paramTypeCombine, unused types)
- `.gitignore` cleaned up (removed Rust/Cargo boilerplate)

## [1.0.0] - 2026-04-11

### Added
- Schema version management with migration support
- Man pages for all commands (auto-generated via cobra/doc)

## [0.3.0] - 2026-04-11

### Added
- `gnb config path` — show active config file path
- `gnb config convert --to <format>` — convert between TOML/YAML/JSON
- `gnb export` / `gnb import` — config portability
- Dual scope: global (`~/.config/ganbatte/`) + project (`.ganbatte.*`)
- `LoadScoped()` with automatic merge (project overrides global)
- Conflict detection between scopes

## [0.2.0] - 2026-04-11

### Added
- Shell detection (zsh/bash/fish)
- History mining — zsh (extended + plain), bash, fish parsers
- `gnb suggest` — frequency-based alias and sequence-based workflow suggestions
- TUI browser with fuzzy search, tag filtering, preview panel
- `gnb run --dry-run` with destructive command detection
- `gnb show` — detailed view of alias/workflow
- `gnb doctor` — environment diagnostics
- Workflow engine: `on_fail` (stop/continue), `confirm` gates, parameter substitution

## [0.1.0] - 2026-04-11

### Added
- Initial release
- `gnb init` — config initialization with format selection
- `gnb add` / `gnb edit` / `gnb remove` — alias CRUD
- `gnb run` — execute alias or workflow
- `gnb list` — list aliases and workflows with tag filtering
- Multi-format config support (TOML/YAML/JSON)
- Workflow engine with ordered steps and parameter substitution
