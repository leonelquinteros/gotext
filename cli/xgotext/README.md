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

## Implementation

This is the first (naive) implementation for this tool. 

It will scan the Go package provided for method calls that matches the method names from the gotext package and write the corresponding translation files to the output directory. 

Isn't able to parse calls to translation functions using parameters inside variables, if the translation string is inside a variable and that variable is used to invoke the translation function, this tool won't be able to parse that string. See this example code: 

```go
// This line will be added to the .po file
gotext.Get("Translate this")

tr := "Translate this string"
// The following line will NOT be added to the .pot file
gotext.Get(tr)
```

The CLI tool traverse sub-directories based on the given input directory.


## Contribute

Please

