# Best Practices for gotext

Localized applications can become complex quickly. Following these best practices will help you keep your project maintainable.

## 1. Organise Your Directory Structure

Consistency is key. Use the standard Gettext directory structure:

```
/locales
  /en_US
    /LC_MESSAGES
      default.po
      errors.po
  /es_ES
    /LC_MESSAGES
      default.po
      errors.po
```

- **LC_MESSAGES**: Always place your `.po` and `.mo` files inside an `LC_MESSAGES` directory or directly under the language code.
- **Simplified Codes**: provide fallbacks (e.g., provide `es` if `es_AR` and `es_ES` share many strings).

## 2. Use Meaningful Domains

Don't put all your translations in a single "default" domain for large apps. Split them logically:
- `ui.po`: Interface elements (buttons, labels).
- `errors.po`: System and error messages.
- `help.po`: Long-form documentation or help text.

## 3. Leverage Context (`msgctxt`)

Avoid ambiguity by using context for short strings that might have different meanings depending on where they appear.

```go
// In code:
gotext.GetC("Open", "File menu")
gotext.GetC("Open", "Lock status")
```

```po
// In PO file:
msgctxt "File menu"
msgid "Open"
msgstr "Abrir"

msgctxt "Lock status"
msgid "Open"
msgstr "Abierto"
```

## 4. Automate String Extraction

Never manually edit `msgid` entries in your `.po` files. Use the `xgotext` CLI tool to scan your code and update your translation files. This ensures your code and translations stay in sync.

```bash
xgotext -p . -o locales/en_US/default.po
```

## 5. Thread Safety

`gotext` is thread-safe. You can safely share a `Locale` object across multiple goroutines (e.g., in an HTTP handler). Avoid re-configuring the global package state (`gotext.Configure`) after your application has started.

## 6. Testing

When writing tests for your application, you can use the `Po` object to load small strings directly, which is often faster than setting up a full directory structure.

```go
po := gotext.NewPo()
po.Parse([]byte("msgid \"test\"\nmsgstr \"prueba\""))
```
