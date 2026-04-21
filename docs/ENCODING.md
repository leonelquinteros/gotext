# Character Encoding with gotext

This guide covers how `gotext` handles character encoding, particularly UTF-8.

## UTF-8 by Default

`gotext` is built to work seamlessly with UTF-8, which is the default encoding for Go source code and strings. We highly recommend using UTF-8 for all your `.po` and `.mo` files.

### Why UTF-8?
1.  **Consistency**: No need for manual conversion between encodings and Go's internal string representation.
2.  **Breadth**: Support for almost any character set in a single file.
3.  **Modern Standard**: UTF-8 is the industry standard for localization.

## Using Other Encodings

While UTF-8 is strongly recommended, the standard GNU Gettext specification allows for other encodings, such as ISO-8859-1.

### Header Configuration
The encoding of a `.po` file is defined in its header:

```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
```

When `gotext` parses these files, it respects the `charset` defined in the `Content-Type` header if possible. However, since Go strings are natively UTF-8, some older encodings may require manual handling or are implicitly converted when read.

## Troubleshooting Encoding Issues

If you see garbled characters or "diamonds" (replacement characters), check the following:

1.  **File Format**: Ensure your `.po` or `.mo` file is actually saved in the encoding specified in its header.
2.  **Terminal/Display**: Ensure your terminal or UI is configured to display the character set you are using.
3.  **Go Source**: Ensure your Go source files are saved as UTF-8 (the default for the Go compiler).

## Recommendations
For the best experience, always:
- Save `.po` files as **UTF-8 (without BOM)**.
- Set the `charset` header to `UTF-8`.
- Use the `xgotext` CLI tool to maintain consistent file formatting.
