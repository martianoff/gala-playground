# [FIXED in 0.15.2] Go Source Files in Subpackages Not Copied to Transpiled Output

## Summary

When an external GALA module contains Go subpackages (e.g., `martianoff/gala-server/httpcore`), the build system adds a `replace` directive for the parent module but not for its Go subpackages. Go's module system requires that subpackages within a `replace`d module are also covered by the replace path, but only if the subpackage is referenced as a separate import.

## Environment

- GALA version: 0.15.0
- OS: Windows 11

## Reproduction

`gala.mod`:

```
module gala-playground

gala 0.15.0

require github.com/martianoff/gala-server 3c657658dcb4
```

`main.gala`:

```gala
package main

import . "martianoff/gala-server"

func main() {
    NewServer().GET("/", (req) => Ok("hello")).Listen()
}
```

The `gala-server` module internally imports its own Go subpackage:

```gala
// gala-server/request.gala
import "martianoff/gala-server/httpcore"
```

Where `httpcore` is a pure Go package (`httpcore/httpcore.go`) within the gala-server module.

Run `gala build .`

## Expected

Build succeeds. The `replace` directive for `martianoff/gala-server` should cover its subpackages automatically (Go module semantics), or the build system should add explicit replace directives for subpackages.

## Actual

```
deps\...\request.gen.go:6:8: package martianoff/gala-server/httpcore is not in std
```

## Generated go.mod

The generated `go.mod` contains:

```
replace martianoff/gala-server => C:/Users/maxmr/.gala/build/.../deps/github.com/martianoff/gala-server@<hash>
```

This should work for Go subpackages — Go's module system resolves `martianoff/gala-server/httpcore` as a subdirectory within the replaced path. The error suggests the directory structure under the replace path may not include the `httpcore/` subdirectory, or the dependency's Go files aren't being included alongside the transpiled `.gen.go` files.

## Dependency structure

```
gala-server/
  server.gala        -> transpiled to server.gen.go
  request.gala       -> transpiled to request.gen.go (imports httpcore)
  response.gala      -> transpiled to response.gen.go
  filter.gala        -> transpiled to filter.gen.go
  httpcore/
    httpcore.go      -> pure Go file (NOT transpiled, must be copied as-is)
```

## Update (0.15.1)

The `httpcore/` directory is now created in the transpiled output, but only `BUILD.bazel` is copied — the actual Go source file `httpcore.go` is missing:

```
deps/.../gala-server@<hash>/httpcore/
  BUILD.bazel     <- copied
  (httpcore.go)   <- MISSING
```

The fix correctly identifies Go subpackages and creates the directory, but the file copy filter excludes `.go` files (or only copies build-system files like `BUILD.bazel`).

## Root Cause

The dependency transpilation step copies `.gala` -> `.gen.go` and `BUILD.bazel` files, but skips raw `.go` source files. Pure Go subpackages like `httpcore/` contain only `.go` files (no `.gala`), so they end up as empty directories with just `BUILD.bazel`.

## Suggested Fix

When processing an external GALA module dependency, copy **all** `.go` files from the module source into the transpiled output directory — not just build-system files. The copy logic should include:
1. All `.go` files at the module root (for mixed GALA + Go modules)
2. All `.go` files in subdirectories (for pure Go subpackages like `httpcore/`)
3. `go.sum` if present (for Go dependency verification)
