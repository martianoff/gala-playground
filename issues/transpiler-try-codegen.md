# [FIXED in 0.16.0] Transpiler: Try() Code Generation Uses Unresolved Generic Type `T`

## Summary

When using `Try()` with pattern matching, the transpiler generates Go code with unresolved generic type parameter `T` instead of the concrete type. Additionally, `Try()` lambdas wrapping Go functions that return `(T, error)` tuples don't handle the error return value.

## Environment

- GALA version: 0.15.3
- OS: Windows 11

---

## Issue 1: Match on Try Uses `T` Instead of Concrete Type

### GALA source

```gala
return Try(() => os.MkdirTemp("", "gala-playground-*")) match {
    case Failure(_) => RunResult(Output = "", Error = "Failed", Time = "")
    case Success(tmpDir) => ...
}
```

### Generated Go (incorrect)

```go
func(obj std.Try[T]) RunResult {
    _tmp := std.Failure[T]{}.Unapply(obj)
    // ...
    var _tmp_12 T
    // ...
}(std.Try{}.Apply(func() string {
    return os.MkdirTemp("", "gala-playground-*")
}))
```

### Errors

```
undefined: T
cannot use generic type std.Try[T any] without instantiation
```

### Expected Generated Go

```go
func(obj std.Try[string]) RunResult {
    _tmp := std.Failure[string]{}.Unapply(obj)
    // ...
    var _tmp_12 string
    // ...
}(std.Try[string]{}.Apply(func() string {
    return os.MkdirTemp("", "gala-playground-*")
}))
```

The transpiler should infer `T = string` from `os.MkdirTemp` return type and instantiate `Try[string]`, `Success[string]`, `Failure[string]` with the concrete type.

---

## Issue 2: Try Lambda Doesn't Handle `(T, error)` Go Return Tuples

### GALA source

```gala
val setup = Try(() => {
    val srcFile = os.Create(filepath.Join(tmpDir, "main.gala"))
    srcFile.WriteString(code)
    srcFile.Close()
    ...
})
```

### Generated Go (incorrect)

```go
var srcFile = std.NewImmutable(os.Create(filepath.Join(tmpDir, "main.gala")))
```

### Errors

```
too many return values: have (string, error), want (string)
too many arguments in call to std.NewImmutable: have (*os.File, error), want (T)
```

### Analysis

`os.MkdirTemp` returns `(string, error)` and `os.Create` returns `(*os.File, error)`. Inside a `Try()` lambda, the transpiler should:
1. Capture both return values: `val result, err = os.MkdirTemp(...)`
2. Check the error and panic if non-nil (Try catches panics): `if err != nil { panic(err) }`
3. Return only the value: `return result`

Or alternatively, generate Go code that destructures the tuple:

```go
std.Try[string]{}.Apply(func() string {
    _v, _err := os.MkdirTemp("", "gala-playground-*")
    if _err != nil { panic(_err) }
    return _v
})
```

---

## Minimal Reproduction

```gala
package main

import "os"

func main() {
    val result = Try(() => os.MkdirTemp("", "test-*"))
    result match {
        case Success(dir) => Println(s"Created: $dir")
        case Failure(err) => Println(s"Error: $err")
    }
}
```

## Suggested Fix

1. When generating match helpers for `Try[T]`, resolve `T` to the concrete type inferred from the matched expression
2. When wrapping Go functions returning `(T, error)` inside `Try()` lambdas, generate error-checking code that panics on error (letting `Try.Apply` catch it as `Failure`)
