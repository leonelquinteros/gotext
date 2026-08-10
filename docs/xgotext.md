# xgotext CLI Tool

`xgotext` extracts translatable strings from Go source code and writes Gettext template (`.pot`) files. It recognizes the getters provided by this module, including calls through imported-package aliases and `gotext.Translator` implementations.

## Installation

```bash
go install github.com/leonelquinteros/gotext/cli/xgotext@latest
```

## Basic usage

Use `-in` to recursively scan one directory tree. One `.pot` file is created per domain in the output directory.

```bash
xgotext -in . -out locales/templates -default default
```

Use `-pkg-tree` when the target is a Go package and its dependencies should also be considered:

```bash
xgotext -pkg-tree . -out locales/templates
```

Exactly one of `-in` and `-pkg-tree` is required.

### Options

| Option | Description |
| --- | --- |
| `-in <directory>` | Recursively scan a source directory. |
| `-pkg-tree <directory>` | Scan a Go package tree, including packages that import gotext. |
| `-out <directory>` | Directory in which to create `<domain>.pot` files. Required. |
| `-default <domain>` | Name used for calls without an explicit domain. Defaults to `default`. |
| `-exclude <directories>` | Comma-separated directory prefixes to skip in `-in` mode. Defaults to `.git`. |
| `-v` | Print paths/packages while they are processed. |

## Supported calls

The extractor supports `Get`, `GetN`, `GetD`, `GetND`, `GetC`, `GetNC`, `GetDC`, and `GetNDC` on the gotext package, locales, PO objects, and `Translator` values. It extracts the string-valued arguments: message ID, plural ID, domain, and context. Numeric plural counts may remain dynamic.

```go
gotext.Get("Save")
gotext.GetN("%d file", "%d files", count)
gotext.GetD("errors", "Unable to save")
gotext.GetC("Open", "menu action")
```

## Constants and variables

String constants are supported, including aliases and constant expressions. Variables are supported when their declaration-time value is statically resolvable and the variable is not assigned a new value before the getter call.

```go
const saveLabel = "Save"
const menuContext = "menu action"

var cancelLabel = "Cancel"

gotext.Get(saveLabel)
gotext.GetC(cancelLabel, menuContext)
```

For correctness, xgotext deliberately skips a mutable variable after it has been reassigned. For example, this does **not** extract either value:

```go
label := "Before"
label = "After"
gotext.Get(label)
```

Calls with dynamic strings—such as values returned from functions, environment variables, maps, or user input—are skipped. Use a constant or an unreassigned, literal-initialized variable when a message must be extracted.

## Workflow

1. Write your messages using gotext getters.
2. Run xgotext to create/update the `.pot` templates.
3. Create language-specific `.po` files from those templates and translate them.
4. Re-run xgotext as source messages change, then merge the updated templates with your translation workflow.

`xgotext` records source locations and de-duplicates matching messages within a domain/context.
