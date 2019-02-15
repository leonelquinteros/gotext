# xgotext

CLI tool to extract translation strings from Go packages into .PO files. 

## Installation

```
go install github.com/leonelquinteros/gotext/cli/xgotext
```

## Usage

```
xgotext /path/to/go/package [/path/to/output/dir]
```

## Implementation

This is the first (naive) implementation for this tool. 

It will scan the Go package provided for method calls that matches the method names from the gotext package and write the corresponding translation files to the output directory. 

Isn't able to parse calls to translation functions using parameters inside variables, if the translation string is inside a variable and that variable is used to invoke the translation function, this tool won't be able to parse that string. See this example code: 

```go
// This line will be added to the .po file
gotext.Get("Translate this")

tr := "Translate this string"
// The following line will NOT be added to the .po file
gotext.Get(tr)
```

The CLI tool doesn't traverse sub-directories and other packages.


## Contribute

Please

