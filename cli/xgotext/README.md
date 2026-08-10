# xgotext

CLI tool to extract translation strings from Go packages into `.pot` files.

## Installation

```
go install github.com/leonelquinteros/gotext/cli/xgotext@latest
```

## Usage

```
Usage of xgotext:
  -default string
        Name of default domain (default "default")
  -exclude string
        Comma separated list of directories to exclude (default ".git")
  -in string
        input dir: /path/to/go/pkg
  -out string
        output dir: /path/to/i18n/files
  -pkg-tree string
        main path: /path/to/go/pkg
  -v    print currently handled directory
```

Exactly one of `-in` and `-pkg-tree` is required; `-out` is always required.

## Details

The tool scans Go source files for method calls that match the getter names from the `gotext` package (`Get`, `GetN`, `GetD`, `GetND`, `GetC`, `GetNC`, `GetDC`, `GetNDC`) and writes the corresponding translation templates (`.pot` files, one per domain) to the output directory.

- `-in` recursively traverses sub-directories down from the given input directory.
- `-pkg-tree` scans a Go package tree, including packages that import gotext.
- `-exclude` skips comma-separated directory prefixes in `-in` mode (defaults to `.git`).
- `-v` prints paths/packages while they are processed.

## Contribute

Please see the project's [Contributing Guidelines](../../CONTRIBUTING.md).

Full usage details, including supported string constants and variables, are documented in the [xgotext guide](../../docs/xgotext.md).