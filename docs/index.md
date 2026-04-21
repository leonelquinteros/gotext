# gotext

GNU gettext utilities for Go.

`gotext` is a native Go implementation of the GNU Gettext utilities. It provides a thread-safe, flexible, and powerful way to handle internationalization (i18n) and localization (l10n) in your Go applications.

## Why use gotext?

- **Native Go**: No external dependencies or CGO required.
- **Thread-safe**: Designed for concurrent use in web servers and high-performance apps.
- **Gettext Compatible**: Supports standard `.po` and `.mo` files, including complex plural forms and contexts.
- **CLI Support**: Includes `xgotext` to automate the extraction of strings from your source code.

## Quick Links

- [Installation](GETTING_STARTED.md#installation)
- [Basic Usage](GETTING_STARTED.md#2-basic-example-package-level-api)
- [CLI Tool](xgotext.md)
- [Plural Forms](PLURALS.md)

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).
