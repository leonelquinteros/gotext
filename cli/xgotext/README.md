# xgotext

CLI tool to extract translation strings from Go packages into .POT files. 

## Installation

```
go install github.com/leonelquinteros/gotext/cli/xgotext
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
```

## Details

It will scan the Go package provided for method calls that matches the method names from the gotext package and write the corresponding translation files to the output directory. 

The CLI tool traverse sub-directories down from the given input directory.


## Contribute

Please

