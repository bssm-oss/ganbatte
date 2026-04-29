# Documentation

This directory contains the longer-form documentation for `ganbatte` (`gnb`). Start with the root [README](../README.md) for the product overview, then use these files when you need details.

## Documents

| Path | Purpose |
|---|---|
| [spec.md](./spec.md) | Product scope, architecture, configuration model, and non-goals |
| [man/](./man) | Generated command reference for every `gnb` subcommand |
| [demo/](./demo) | Demo assets used by the README |

## Common Tasks

### Install and Verify

```bash
brew install --cask bssm-oss/tap/ganbatte
gnb doctor
gnb --help
```

or:

```bash
go install github.com/bssm-oss/ganbatte/cmd/gnb@latest
gnb doctor
```

### Generate Man Pages

```bash
go run cmd/gendoc.go
```

Generated files are written to [man/](./man). Regenerate them after changing command names, flags, examples, or descriptions.

### Validate a Release Candidate

```bash
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
goreleaser check
goreleaser release --snapshot --clean
```

Before calling a release done, verify at least one clean consumer path:

```bash
go install github.com/bssm-oss/ganbatte/cmd/gnb@<version>
brew install --cask bssm-oss/tap/ganbatte --force
```

## Configuration Files

`ganbatte` supports three equivalent config formats:

- TOML: `config.toml` / `.ganbatte.toml`
- YAML: `config.yaml` / `.ganbatte.yaml`
- JSON: `config.json` / `.ganbatte.json`

Global config is stored under `~/.config/ganbatte/`. Project config is stored as `.ganbatte.*` in a repository and can be discovered from nested directories.

## Scope Boundaries

`gnb` intentionally does not provide machine sync, remote workflow execution, Make/just replacement semantics, AI command generation, or a plugin system. See [spec.md](./spec.md) for the full non-goals list.
