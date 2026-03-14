do not# Transpiler Code Generation Errors (0.16.0)

## Summary

After fixes in 0.16.0, several Go compilation errors were found. Most fixed in 0.16.1, some remain.

## Environment

- GALA version: 0.16.0 (original), retested on 0.16.1+
- OS: Windows 11

---

## Issue 1: `(T, error)` Not Destructured Inside Multi-Statement Try Lambda

**Status: FIXED**

Go functions returning `(T, error)` inside a multi-statement `Try()` lambda block are not destructured when assigned to a `val`.

### GALA source

```gala
val setup = Try(() => {
    val srcFile = os.Create(filepath.Join(tmpDir, "main.gala"))
    srcFile.WriteString(code)
    ...
})
```

### Generated Go (incorrect)

```go
var srcFile = std.NewImmutable(os.Create(filepath.Join(tmpDir, "main.gala")))
```

### Error

```
too many arguments in call to std.NewImmutable: have (*os.File, error), want (T)
```

### Analysis

The `(T, error)` destructuring fix from 0.15.3 works for `Try(() => os.MkdirTemp(...))` (single-expression lambda) but NOT for `val` assignments inside a multi-statement `Try()` lambda block. The transpiler needs to apply the same `_v, _err := ...; if _err != nil { panic(_err) }` pattern for `val` assignments inside Try blocks.

### Workaround

Use `val x, err = f()` with explicit destructuring inside the Try block:

```gala
val setup = Try(() => {
    val srcFile, _ = os.Create(filepath.Join(tmpDir, "main.gala"))
    srcFile.WriteString(code)
    srcFile.Close()
    return true
})
```

See also: `transpiler-try-chained-method.md` for the related chained-method variant.

---

## Issue 2: Missing `io/fs` Import for `os.Stat` Return Type

**Status: FIXED in 0.16.1**

The transpiler now correctly adds `import "io/fs"` when `os.Stat` return type resolves to `fs.FileInfo`.

---

## Issue 3: `unit` Type Undefined in Match on Void Functions

**Status: FIXED in 0.16.1**

Void match expressions are now correctly generated as if-else statements.
