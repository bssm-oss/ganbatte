# ganbatte (`gnb`)

[한국어](./README.ko.md) | [Command reference](./docs/man) | [Specification](./docs/spec.md) | [Releases](https://github.com/bssm-oss/ganbatte/releases)

> Your shell history knows what you repeat. ganbatte turns it into aliases and workflows.

Your `.zshrc` knows your shortcuts. Your shell history knows your habits. Your project knows its workflows.

`ganbatte` puts them in one place: a local command hub for migrating aliases, mining repeated commands, and promoting command sequences into reusable workflows.

```bash
brew install bssm-oss/tap/ganbatte

gnb suggest   # find repeated commands and workflow candidates
gnb migrate   # import existing shell aliases
gnb           # browse everything in a TUI
```

![ganbatte TUI demo](docs/demo/tui.gif)

## Why It Exists

Shell aliases are fast until your `.zshrc` becomes an alias graveyard. Make, just, and task are great for project build tasks, but they are not designed for personal command shortcuts, cross-shell alias migration, or history-driven recommendations.

Atuin remembers what you typed. `ganbatte` turns what you repeat into commands you can keep.

`gnb` sits in the middle:

| Need | Shell aliases | make/just/task | ganbatte |
|---|---|---|---|
| Personal shortcuts | Yes | No | Yes |
| Project onboarding workflows | Manual | Yes | Yes |
| Cross-shell config | No | Partial | Yes |
| Parameters without shell functions | No | Yes | Yes |
| Dangerous command confirmation | No | Manual | Yes |
| TUI discovery | No | Usually no | Yes |
| History-based suggestions | No | No | Yes |

## Killer Features

### 1. Suggest Shortcuts From Real Usage

`gnb shell-init` can passively append command history to `~/.local/share/ganbatte/track.log` without spawning the `gnb` binary for every command. `gnb suggest` then uses shell history or the track log to recommend aliases, parameterized aliases, and workflows.

```bash
gnb suggest
gnb suggest --apply
gnb suggest --from-history
```

```text
$ gnb suggest
Analyzing ganbatte track log (/Users/you/.local/share/ganbatte/track.log) (312 entries)...

=== Alias Suggestions ===
  1. c = claude
     Used 5 times · saves ~25 keystrokes

=== Parameterized Alias Suggestions ===
  1. gcl(repo) -> git clone {repo}
     Pattern 'git clone <...>' used 25 times with 25 variants

=== Workflow Suggestions ===
  1. git-add
     Step 1: git add .
     Step 2: git commit -m "update"
     Step 3: git push
     Sequence appeared 7 times

Applying all suggestions would save ~89 keystrokes based on your history.
```

Suggestions are ranked by estimated keystrokes saved, not just raw frequency. Destructive commands detected during `--apply` are marked with `confirm = true`.

![gnb suggest demo](docs/demo/suggest.gif)

### 2. Migrate Existing Shell Aliases

Bring your current `.zshrc`, `.bashrc`, `.bash_aliases`, or fish aliases into one managed config.

```bash
gnb migrate
gnb migrate --dry-run
gnb migrate --shell zsh
```

Example flow:

```text
$ gnb migrate
Found 17 aliases in /Users/you/.zshrc
Found 8 aliases in /Users/you/.bash_aliases

25 new alias(es) to import:
  gs = "git status -sb"
  ll = "eza -alF --git"
  dc = "docker compose"

Import all? [Y/n] y
```

![gnb migrate demo](docs/demo/migrate.gif)

### 3. Browse Everything in a TUI

Run `gnb` with no arguments to open the command hub: fuzzy search, preview, tags, and global/project labels for every alias and workflow.

![ganbatte TUI demo](docs/demo/tui.gif)

## Privacy

`gnb suggest` analyzes local shell history files and the local ganbatte track log. Nothing is sent to a server, and ganbatte has no telemetry, account, cloud sync, or AI command generation.

## Install

### Homebrew

```bash
brew install bssm-oss/tap/ganbatte
```

### Go

```bash
go install github.com/bssm-oss/ganbatte/cmd/gnb@latest
```

### Release Archive

Download a platform archive from [GitHub Releases](https://github.com/bssm-oss/ganbatte/releases), extract `gnb`, and place it on your `PATH`.

## Quick Start

### Start From Existing Aliases

```bash
gnb migrate
echo 'eval "$(gnb shell-init)"' >> ~/.zshrc
source ~/.zshrc

gs           # runs the migrated alias
dc up -d     # works like a native shell function
```

For bash, add the same `eval` line to `~/.bashrc`. For fish:

```fish
gnb shell-init | source
```

### Start From Scratch

```bash
gnb init
gnb add gs "git status -sb"
gnb run gs
```

Run `gnb` with no arguments to open the TUI browser.

## Core Features

### Global and Project Scopes

Global config lives in `~/.config/ganbatte/config.{toml,yaml,json}`. Project config lives in `.ganbatte.{toml,yaml,json}` in the current directory or a trusted parent repository.

Project entries override global entries with the same name. `gnb run` asks for confirmation before executing a project item that shadows a global item.

```bash
gnb list --scope global
gnb list --scope project
gnb run setup --yes
```

### Parameterized Aliases

```toml
[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.glog]
cmd = "git log --oneline -{count}"
params = ["count"]
default_params = { count = "10" }
```

```bash
gco feature/login
glog
glog 30
```

### Workflows

```toml
[workflow.deploy]
description = "Lint, test, build, and push"
params = ["branch"]
tags = ["deploy", "ci"]

[[workflow.deploy.steps]]
run = "pnpm lint"

[[workflow.deploy.steps]]
run = "pnpm test"
on_fail = "stop"

[[workflow.deploy.steps]]
run = "git push origin {branch}"
confirm = true
```

```bash
gnb run deploy main --dry-run
gnb run deploy main
```

### Confirm Guardrails

Use `confirm = true` for commands that should never run accidentally.

```toml
[alias.clean-docker]
cmd = "docker system prune -af"
confirm = true
```

### TUI Browser

```bash
gnb
```

The TUI gives you fuzzy search, preview, tag filtering, and global/project labels for aliases and workflows.

### Multi-Format Config

TOML, YAML, and JSON are first-class and equivalent.

```bash
gnb init --format toml
gnb config convert --to yaml
gnb config path
```

## Command Overview

| Command | Purpose |
|---|---|
| `gnb` | Open the TUI browser |
| `gnb init` | Create a global or project config |
| `gnb add <name> <cmd>` | Add an alias |
| `gnb edit <name> <cmd>` | Update an alias |
| `gnb remove <name>` | Remove an alias or workflow |
| `gnb list` | List aliases and workflows |
| `gnb show <name>` | Show details |
| `gnb run <name> [args...]` | Run an alias or workflow |
| `gnb migrate` | Import shell aliases |
| `gnb suggest` | Recommend aliases/workflows from history |
| `gnb shell-init` | Print shell integration functions |
| `gnb doctor` | Diagnose setup issues |
| `gnb export` / `gnb import` | Move config between machines |

For full command docs, see [docs/man](./docs/man).

## Configuration Reference

```toml
version = "0.1.0"

[alias.gs]
cmd = "git status -sb"
tags = ["git"]

[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.nuke]
cmd = "git reset --hard HEAD"
confirm = true

[workflow.test]
description = "Run checks"
tags = ["ci"]

[[workflow.test.steps]]
run = "go test ./..."

[[workflow.test.steps]]
run = "go vet ./..."
```

## Documentation

- [Documentation index](./docs/README.md)
- [Full project specification](./docs/spec.md)
- [Generated man pages](./docs/man)
- [Contributing guide](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)

## Development

```bash
go build -o gnb .
go test ./...
go test -race ./...
go vet ./...
go run cmd/gendoc.go
```

Release builds are handled by GoReleaser on pushed `v*` tags. Consumer paths verified for the current release include GitHub release archives, `go install github.com/bssm-oss/ganbatte/cmd/gnb@v1.5.6`, and the Homebrew formula.

## Non-Goals

`ganbatte` is not a machine sync service, a remote execution system, a Make/just replacement, an AI command generator, or a plugin platform.

## License

MIT
