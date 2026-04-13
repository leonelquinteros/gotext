# xgotext CLI Tool

`xgotext` is a command-line tool designed to help you extract translatable strings from your Go source code and generate or update Gettext PO files.

## 1. Installation

Install `xgotext` using `go install`:

```bash
go install github.com/leonelquinteros/gotext/cli/xgotext@latest
```

## 2. Usage

To extract strings from your project and create a new PO file:

```bash
xgotext -p . -o locales/en_US/default.po
```

### Options:
- `-p <path>`: The directory path to scan for Go files (default: current directory).
- `-o <output>`: The output path for the generated PO file.
- `-d <domain>`: The domain to extract (default: "default").
- `-k <keyword>`: Add custom keywords to look for (default: `Get`, `GetD`, `GetN`, `GetND`, `GetC`, `GetDC`, `GetNC`, `GetNDC`).

### 3. Example Workflow

1.  **Write your Go code** using `gotext.Get("Hello!")`.
2.  **Run `xgotext`** to generate the initial `en_US/default.po` file.
3.  **Translate** the PO file into other languages (e.g., `es_AR/default.po`).
4.  **Update** your translations later as your code changes by re-running `xgotext` with the same output path.

## 4. How it works

`xgotext` parses your Go files looking for function calls that match the default keywords or any custom ones you've specified. It then collects all unique `msgid` and `msgctxt` pairs and writes them to the specified output file, preserving any existing translations if the file already exists.
