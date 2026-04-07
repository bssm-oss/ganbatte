# PRD: ganbatte MVP (v0.1 "It works")

## Introduction

ganbatte is a workflow/shortcut management CLI tool designed for lazy developers who want to minimize finger movements while maximizing productivity. The MVP focuses on providing core alias and workflow management capabilities with multi-format config support, allowing users to replace manual command typing with simple shortcuts.

## Goals

- Provide a single binary (`gnb`) for managing shell aliases and workflows
- Support equal functionality across TOML, YAML, and JSON config formats
- Enable basic CRUD operations for aliases and workflows
- Support parameter substitution in workflows
- Provide global and project-scoped configuration
- Include initialization wizard for easy setup
- Generate shell completion scripts
- Provide basic error handling and help documentation

## User Stories

### US-001: Initialize ganbatte configuration
**Description:** As a new user, I want to initialize ganbatte with a configuration wizard so I can start using it quickly without manual setup.

**Acceptance Criteria:**
- [ ] `gnb init` command starts an interactive wizard
- [ ] Wizard detects user's shell (bash/zsh/fish)
- [ ] Wizard asks for preferred config format (TOML/YAML/JSON)
- [ ] Wizard creates default config file with example aliases/workflows
- [ ] Wizard outputs success message with next steps
- [ ] Typecheck passes

### US-002: Add a new alias
**Description:** As a user, I want to add new command aliases so I can shorten frequently used commands.

**Acceptance Criteria:**
- [ ] `gnb add <name> <command>` creates a new alias
- [ ] Alias is saved in the active configuration file
- [ ] Duplicate alias names are rejected with helpful error message
- [ ] Alias can be viewed with `gnb list`
- [ ] Typecheck passes

### US-003: List all aliases and workflows
**Description:** As a user, I want to see all my configured aliases and workflows so I know what shortcuts are available.

**Acceptance Criteria:**
- [ ] `gnb list` shows all aliases with their commands
- [ ] `gnb list` shows all workflows with their descriptions
- [ ] Output is clearly formatted and easy to read
- [ ] Empty state shows appropriate message when no items exist
- [ ] Typecheck passes

### US-004: Edit an existing alias
**Description:** As a user, I want to modify existing aliases so I can update them when my command preferences change.

**Acceptance Criteria:**
- [ ] `gnb edit <name> <new-command>` updates an existing alias
- [ ] Editing non-existent alias shows helpful error message
- [ ] Updated alias is immediately available for use
- [ ] Typecheck passes

### US-005: Remove an alias or workflow
**Description:** As a user, I want to delete aliases and workflows I no longer need so my configuration stays clean.

**Acceptance Criteria:**
- [ ] `gnb remove <name>` deletes the specified alias or workflow
- [ ] Removing non-existent item shows helpful error message
- [ ] Removed item no longer appears in `gnb list`
- [ ] Typecheck passes

### US-006: Run an alias
**Description:** As a user, I want to execute my aliases so I can run complex commands with minimal typing.

**Acceptance Criteria:**
- [ ] `gnb run <alias-name>` executes the associated command
- [ ] Command output is displayed to the user
- [ ] Errors from command execution are properly propagated
- [ ] Typecheck passes

### US-007: Create a workflow with parameters
**Description:** As a user, I want to define workflows with parameters so I can create reusable command sequences.

**Acceptance Criteria:**
- [ ] Workflows can be defined with named parameters like `{branch}`
- [ ] Workflow steps can include conditional execution logic
- [ ] Workflow steps can require confirmation before execution
- [ ] Typecheck passes

### US-008: Run a workflow with parameter substitution
**Description:** As a user, I want to execute workflows with parameter substitution so I can customize workflows for different contexts.

**Acceptance Criteria:**
- [ ] `gnb run <workflow-name> <param-values>` substitutes parameters correctly
- [ ] Multiple parameters are handled in order
- [ ] Missing parameters default to empty strings
- [ ] Typecheck passes

### US-009: Generate shell completion
**Description:** As a user, I want shell completion for gnb commands so I can use tab completion for faster interaction.

**Acceptance Criteria:**
- [ ] `gnb completion <shell>` outputs valid completion script
- [ ] Supported shells: bash, zsh, fish
- [ ] Generated completion works when sourced in shell
- [ ] Typecheck passes

### US-010: Global and project scope support
**Description:** As a user, I want configuration to work both globally and per-project so I can have both personal shortcuts and project-specific shortcuts.

**Acceptance Criteria:**
- [ ] Global config is loaded from `~/.config/ganbatte/config.*`
- [ ] Project config is loaded from `./ganbatte.*` or `./.ganbatte.*`
- [ ] Project-scoped items override global items with same name
- [ ] `gnb list` shows which scope each item comes from (future enhancement)
- [ ] Typecheck passes

## Functional Requirements

- FR-1: Implement `gnb init` command for initial setup
- FR-2: Implement `gnb add <name> <command>` for adding aliases
- FR-3: Implement `gnb list` for displaying aliases and workflows
- FR-4: Implement `gnb edit <name> <command>` for modifying aliases
- FR-5: Implement `gnb remove <name>` for deleting aliases/workflows
- FR-6: Implement `gnb run <name> [args...]` for executing aliases/workflows
- FR-7: Implement `gnb completion <shell>` for shell autocompletion
- FR-8: Support TOML, YAML, and JSON config formats with equal functionality
- FR-9: Implement parameter substitution in workflows (`{param}` syntax)
- FR-10: Support global (`~/.config/ganbatte/`) and project (`./ganbatte.*`) scopes
- FR-11: Provide basic error handling with informative messages
- FR-12: Implement `--help` flag for all commands

## Non-Goals

- History mining/suggestion features (planned for v0.2)
- TUI browser interface (planned for v0.2)
- Dry-run execution mode (planned for v0.2)
- Conditional workflow steps beyond basic implementation (planned for v0.2)
- Tag-based filtering (planned for v0.2)
- Machine synchronization (delegated to syncingsh)
- Remote workflow execution
- AI-based command generation
- Plugin system (planned for v1.x+)

## Technical Considerations

- Use Cobra for CLI command structure
- Use Viper for multi-format configuration handling
- Configuration files stored in TOML by default but support YAML/JSON
- Global configuration in `$HOME/.config/ganbatte/`
- Project configuration in current directory (`.ganbatte.*`)
- Internal representation (IR) isolates format-specific differences
- Error handling uses standard Go error wrapping with `fmt.Errorf`
- No global state; configuration passed explicitly between components

## Success Metrics

- User can replace daily `lazyasf` usage with ganbatte for one week without inconvenience
- All core CRUD operations work reliably across all three config formats
- Parameter substitution works correctly in workflows
- Shell completion generates functional scripts for bash/zsh/fish
- Error messages are clear and actionable
- Code maintains <10% cyclomatic complexity per function
- Test coverage >75% for core packages (config, workflow)

## Open Questions

- Should project config override global config or merge with it?
- What should be the exact precedence order when multiple config formats exist in same directory?
- Should the init wizard create example workflows in addition to aliases?
- How should we handle configuration file permissions for security?