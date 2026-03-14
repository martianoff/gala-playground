# [FIXED] Feature Request: `Option[T].OrElse()` Method

## Summary

`Option[T]` is missing an `OrElse` method, which is standard in functional languages (Scala, Rust, Haskell). This forces awkward workarounds when chaining optional lookups.

## Environment

- GALA version: 0.16.1+
- OS: Windows 11

## Expected API

```gala
func (o Option[T]) OrElse(alternative Option[T]) Option[T]
```

Returns the option if it's `Some`, otherwise returns `alternative`.

## Use Case

Chaining multiple fallback lookups:

```gala
// DESIRED: clean fallback chain
func findGala() Option[string] =
    candidates.Find((c) => Try(() => os.Stat(c)).IsSuccess())
        .OrElse(Try(() => exec.LookPath("gala")).ToOption())
        .OrElse(Try(() => exec.LookPath("gala.exe")).ToOption())
```

## Current Workaround

Use `GetOrElse("")` to extract with a sentinel, then wrap back in `Option` with `Filter`:

```gala
// WORKAROUND: sentinel + filter roundabout
val path = candidates
    .Find((c) => Try(() => os.Stat(c)).IsSuccess())
    .GetOrElse(
        Try(() => exec.LookPath("gala"))
            .OrElse(Try(() => exec.LookPath("gala.exe")))
            .GetOrElse(""))

return Some(path).Filter((p) => p != "")
```

This is error-prone (sentinel `""` could be a valid value for other types) and violates GALA's own lint rule against sentinel return values.

## Note

`Try[T]` already has `OrElse` — adding it to `Option[T]` would make the API consistent.
