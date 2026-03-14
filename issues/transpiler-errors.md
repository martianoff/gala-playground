# [FIXED in 0.15.3] Transpiler Code Generation Errors (0.15.2)

## Summary

After fixing the module fetch and build system issues, `gala build` now transpiles successfully but the generated Go code has 4 compilation errors.

## Environment

- GALA version: 0.15.2
- OS: Windows 11

---

## Issue 1: `std.NewEmbeddedFS` Undefined

### Generated Go code (line 25)

```go
var _embed_exampleFiles embed.FS
var exampleFiles = std.NewEmbeddedFS(_embed_exampleFiles)
```

### Error

```
undefined: std.NewEmbeddedFS
```

### GALA source

```gala
embed val exampleFiles EmbeddedFS = "examples/*"
```

### Analysis

The transpiler generates a call to `std.NewEmbeddedFS()` but this function may not be exported in the stdlib at `~/.gala/stdlib/v0.15.2/std/`. The function exists in `std/embedded_fs.go` in the GALA repo but might not be included in the extracted stdlib.

### Suggested Fix

Verify that `embedded_fs.go` is included in the stdlib extraction step and that `NewEmbeddedFS` is exported.

---

## Issue 2: ~~`Array.ToSlice()` Undefined~~ — SOURCE CODE FIX

**Not a transpiler bug.** The correct method is `Array.ToGoSlice()`, not `.ToSlice()`. Fixed in source.

---

## Issue 3: `runtime.GOOS` Used as a Type in Match

### Generated Go code (line 246)

```go
var cmd = std.NewImmutable(func(obj runtime.GOOS) *exec.Cmd {
    if obj == "windows" {
```

### Error

```
runtime.GOOS (constant "windows" of type string) is not a type
```

### GALA source

```gala
val goos = runtime.GOOS
val cmd = goos match {
    case "windows" => exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
    case "darwin" => exec.Command("open", url)
    case _ => exec.Command("xdg-open", url)
}
```

### Analysis

The transpiler generates a match helper as an immediately-invoked function `func(obj runtime.GOOS)`. It uses the **value expression** `runtime.GOOS` as the **parameter type**. Since `runtime.GOOS` is a Go string constant, not a type, this is invalid Go.

The transpiler should infer the type of `goos` (which is `string`) and generate `func(obj string)` instead.

### Reproduction (minimal)

```gala
package main

import "runtime"

func main() {
    val goos = runtime.GOOS
    val msg = goos match {
        case "windows" => "win"
        case _ => "other"
    }
    Println(msg)
}
```

### Suggested Fix

When generating the match helper function, use the **inferred type** of the matched expression, not the expression text itself, as the parameter type.

---

## Issue 4: Lambda Parameters Typed as `any` Instead of `Request`

### Generated Go code (lines 280-283)

```go
.POST("/api/run", func(req any) Response {
    return handleRun(galaBin.Get(), req)
})
```

### Error

```
cannot use func(req any) Response as func(server.Request) server.Response value
cannot use req (variable of interface type any) as server.Request value
```

### GALA source

```gala
val server = NewServer().
    POST("/api/run", (req) => handleRun(galaBin, req)).
    GET("/api/version", (req) => handleVersion(galaBin, req))
```

### Analysis

The lambdas `(req) => handleRun(galaBin, req)` have implicit parameter types. The transpiler should infer `req` as `Request` from the `.POST()` method signature:

```gala
func (s Server) POST(pattern string, handler func(Request) Response, filters ...Filter) Server
```

Instead, it generates `func(req any) Response` — using `any` as a fallback when it can't resolve the type.

### Reproduction (minimal)

```gala
package main

import . "martianoff/gala-server"

func greet(req Request) Response = Ok("hello")

func main() {
    NewServer().
        GET("/", (req) => greet(req)).
        Listen()
}
```

### Suggested Fix

When a lambda with implicit parameter types is passed as an argument to a function/method, infer the parameter types from the corresponding function parameter's type signature. In this case, `.POST()` expects `func(Request) Response`, so `(req) =>` should resolve `req` to `Request`.
