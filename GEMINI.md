# GEMINI.md - Gemini CLI Instructions for gotext

> **Important**: Read and follow [`AGENTS.md`](./AGENTS.md) first. It is the source of truth for project architecture, conventions, and development workflows. The instructions below are specific to Gemini CLI and supplement — not replace — those guidelines.

## Gemini-Specific Instructions

- Always work in a feature branch. Never commit directly to `master`.
- Before submitting, verify changes with existing and new tests (`go test ./...`).
- Propose a PR to `master` after confirmation, using the `gh` CLI when available.
