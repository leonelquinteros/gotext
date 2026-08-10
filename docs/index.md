# gotext

GNU gettext utilities for Go.

`gotext` is a native Go implementation of the GNU Gettext utilities. It provides a thread-safe, flexible, and powerful way to handle internationalization (i18n) and localization (l10n) in your Go applications.

## Why use gotext?

- **Native Go**: The core library has no external dependencies and requires no CGO.
- **Thread-safe**: Designed for concurrent use in web servers and high-performance apps.
- **Gettext Compatible**: Supports standard `.po` and `.mo` files, including complex plural forms and contexts.
- **CLI Support**: Includes `xgotext` to automate the extraction of strings from your source code.

## Guides

- [Getting Started](GETTING_STARTED.md) — install, first app, and the `Locale` object.
- [Plural Forms](PLURALS.md) — complex pluralization rules.
- [xgotext CLI](xgotext.md) — extract translatable strings from your source code.
- [Character Encoding](ENCODING.md) — UTF-8 handling and legacy encodings.
- [Advanced Usage](ADVANCED.md) — `embed.FS`, named formatting, caching, custom backends.
- [Best Practices](BEST_PRACTICES.md) — keep your translations maintainable.

## Quick Links

- [Installation](GETTING_STARTED.md#1-installation)
- [Basic Example](GETTING_STARTED.md#2-basic-example-package-level-api)
- [API Reference](https://pkg.go.dev/github.com/leonelquinteros/gotext)

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).