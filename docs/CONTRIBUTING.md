# CONTRIBUTING

This open source project welcomes everybody that wants to contribute to it by implementing new features, fixing bugs, testing, creating documentation or simply talk about it.

Most contributions will start by creating a new Issue to discuss what the contribution is about and to agree on the steps to move forward.

## Issues

All issues reports are welcome. Open a new Issue whenever you want to report a bug, request a change or make a proposal.

This should be your start point of contribution.

## Pull Requests

If you have any changes that can be merged, feel free to send a Pull Request.

Usually, you'd want to create a new Issue to discuss about the change you want to merge and why it's needed or what it solves.

## Development Workflow

The project's architecture, conventions, and development workflows are documented in the repository's `AGENTS.md` file, which is the single source of truth for contributors (human or AI agent). Key rules:

- **Branching**: All development happens in dedicated feature branches (e.g., `feature/description-of-change`, `fix/bug-description`).
- **Target branch**: All Pull Requests must be made against `master`. No direct commits to `master`.
- **Testing**: Every new feature or bug fix must include corresponding tests. Use the `fixtures/` directory for sample PO/MO files.

### Pre-Submit Checklist

Before submitting a Pull Request, verify:

1. All existing tests pass: `go test ./...`
2. Code passes static analysis: `go vet ./...`
3. Code is properly formatted: `gofmt`
4. New features or bug fixes include corresponding tests.
5. Commit messages are clear and descriptive.