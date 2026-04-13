# GEMINI.md - gotext Project Context & Guidelines

This document serves as the primary source of truth for the `gotext` project architecture, conventions, and workflows. Adherence to these guidelines is mandatory for all development tasks.

## Project Overview
`gotext` is a GNU Gettext implementation for Go (Golang). it provides a thread-safe, hierarchical translation system that supports PO/MO files, plural forms, and context-aware translations.

## Core Architecture
The project follows a hierarchical structure for managing translations:
1. **Global Package API (`gotext.go`)**: Provides convenient, global functions (e.g., `Get`, `GetD`, `GetN`) that rely on a shared `globalConfig`.
2. **Locales (`locale.go`)**: Represent a specific language (e.g., "en_US", "es_ES"). A `Locale` container manages multiple `Domains`.
3. **Domains (`domain.go`)**: The core translation engine for a specific set of messages. It handles:
    - Message retrieval (with support for plural forms and `msgctxt`).
    - Header parsing (including plural-form rules).
    - Thread-safety via `sync.RWMutex`.
4. **Translators (`translator.go`)**: An interface for translation sources.
    - **PO (`po.go`)**: Parsers for human-readable Gettext Portable Object files.
    - **MO (`mo.go`)**: Parsers for binary Gettext Machine Object files.
5. **Plural Forms (`plurals/`)**: A dedicated package for compiling and evaluating GNU Gettext plural form expressions.

## CLI Tools
- **xgotext (`cli/xgotext/`)**: A tool to scan Go source files, extract translatable strings (matching `Get`, `GetD`, etc.), and generate/update PO files.

## Development Workflows
### Branching & PRs
- **Feature Branches**: All development MUST happen in dedicated feature branches (e.g., `feature/description-of-change`).
- **Target Branch**: All Pull Requests MUST be made against the `master` branch.
- **Review Requirement**: No direct commits to `master` are allowed. All changes must be reviewed via a PR.
- **GitHub Integration**: If the `gh` CLI tool is installed and authenticated, it should be used for managing Pull Requests, Issues, and Reviews. Otherwise, these tasks must be performed manually through the GitHub web interface.

### Coding Standards
- **Idiomatic Go**: Follow standard Go conventions (`gofmt`, `go vet`).
- **Thread Safety**: Ensure all shared state (like in `Domain`) is protected by appropriate synchronization primitives.
- **Testing**:
    - Every new feature or bug fix must include corresponding tests.
    - Use the `fixtures/` directory for sample PO/MO files.
    - Run `go test ./...` to ensure no regressions.

## Key Files & Symbols
- `gotext.go`: `Configure`, `Get`, `GetD`, `GetN`.
- `locale.go`: `Locale`, `NewLocale`, `AddDomain`.
- `domain.go`: `Domain`, `Get`, `GetN`, `parseHeaders`.
- `translator.go`: `Translator` interface.
- `plurals/compiler.go`: `Compile` function for plural expressions.

## Instructions for Gemini CLI
- Always work in a feature branch.
- Before submitting, verify changes with existing and new tests.
- Propose a PR to `master` after confirmation.
