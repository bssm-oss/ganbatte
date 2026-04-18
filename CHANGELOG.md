# Changelog

All notable changes to this project will be documented in this file.

## [1.4.1] - 2026-04-18

### Added
- `gnb suggest --apply` interactive mode: prompts `[y/N/q]` per suggestion, `q` quits immediately
- `gnb doctor` now shows passive tracking status — entry count, how many more needed before `gnb suggest` uses track.log, and log path
- Workflow names changed from initials-based (`wf-ga`) to readable `verb-noun` format (`git-add`, `npm-run`, `docker-build`)
- 30-minute time-gap filter: ignores command sequences that span more than 30 minutes, eliminating cross-session false positives

### Fixed
- `TestLogPath_RespectsXDG` was failing on Windows due to hardcoded `/` path separator — now uses `filepath.Join`

## [1.3.0] - 2026-04-18

### Added
- **Passive shell tracking**: `gnb shell-init` now embeds tracking hooks that append command history to `~/.local/share/ganbatte/track.log` without spawning the gnb binary — zero latency impact
  - zsh: `preexec_functions` + `precmd_functions`
  - bash: `trap DEBUG` + `PROMPT_COMMAND`
  - fish: `--on-event fish_postexec`
- **`gnb suggest` — parameterized alias detection**: finds command groups with a shared N-token prefix and varying argument position, with three quality gates:
  - Gate 1: subcommand discrimination — rejects positions where tokens follow the varying one
  - Gate 2: argument-likeness — rejects when varying tokens look like subcommand keywords
  - Gate 3: flag rejection — rejects groups where most varying tokens are flags
- **`gnb suggest` — keystroke impact scoring**: suggestions sorted by `(len(cmd) - len(alias)) × frequency` instead of raw frequency
- **`gnb suggest --from-history`**: force shell history source, skip track.log
- **Universal alias collision prevention**: 20 ecosystem-claimed names (`gc`, `gs`, `gco`, `ll`, `k`, ...) excluded from name generation with automatic longer fallbacks
- **Destructive command auto-confirm**: `rm`, `kill`, `git reset`, `git push --force`, etc. get `confirm = true` automatically on `--apply`
- **Noise filter**: multi-line continuations (`\`), comments (`#`), all-punctuation strings, and commands under 4 chars are excluded from analysis
- **Track log rotation**: `track.log` auto-rotates to `track.log.1` when it exceeds 10 MB
- `internal/track` package: `LogPath()`, `Parse()`, `Count()` (XDG-aware, cheap line counter)

## [1.2.0] - 2026-04-17

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
