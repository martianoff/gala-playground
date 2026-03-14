# [FIXED in 0.15.0] Generated go.mod Missing require and replace for External GALA Module Dependencies

## Summary

When a GALA project depends on an external module (e.g., `gala-server`), the generated `go.mod` in the build workspace has a `replace` directive for `github.com/martianoff/gala-server` but the generated Go code imports `martianoff/gala-server` (without `github.com/` prefix). The `require` directive for the dependency is also missing.

## Environment

- GALA version: 0.14.0
- OS: Windows 11

## Reproduction

`gala.mod`:

```
module gala-playground

gala 0.14.0

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

Run `gala build .`

## Expected

The generated `go.mod` should contain:

```
require martianoff/gala-server v0.0.0

replace martianoff/gala-server => /path/to/deps/gala-server@<hash>
```

Matching the import path used in the generated Go code (`martianoff/gala-server`).

## Actual

Generated `go.mod` has:

```
replace github.com/martianoff/gala-server => C:/Users/maxmr/.gala/build/.../deps/github.com/martianoff/gala-server@<hash>
```

But:
1. No `require` directive for `martianoff/gala-server`
2. The `replace` uses `github.com/martianoff/gala-server` while the generated Go code imports `martianoff/gala-server`

Error:

```
package martianoff/gala-server is not in std
```

## Context: Importing a GALA Module

This is the first external GALA module dependency (outside the stdlib). The full workflow to add and use an external module:

1. **Add dependency**: `gala mod add github.com/martianoff/gala-server@<commit>`
2. **gala.mod updated**:
   ```
   require github.com/martianoff/gala-server 3c657658dcb4
   ```
3. **Module fetched** to `~/.gala/pkg/mod/github.com/martianoff/gala-server@<hash>/`
4. **Import in .gala source** uses the module's internal path (from its own `gala.mod`):
   ```gala
   import . "martianoff/gala-server"
   ```
   The gala-server's `gala.mod` declares `module martianoff/gala-server`, so the import path is `martianoff/gala-server` — not the GitHub URL.

5. **`gala build`** should:
   - Transpile the dependency's `.gala` files (already done — stored in `~/.gala/build/<hash>/deps/`)
   - Add a `require` + `replace` pair to the generated `go.mod` using the module's internal import path
   - The dependency's own transitive dependencies (e.g., `martianoff/gala/concurrent`) should also get `require` + `replace` entries

Currently step 5 fails because:
- The `require` for `martianoff/gala-server` is missing entirely
- The `replace` directive uses the GitHub URL (`github.com/martianoff/gala-server`) instead of the module's internal path (`martianoff/gala-server`)
- The dependency's transitive requires (e.g., `martianoff/gala/concurrent` used by gala-server's filter.gala) are not added

### Dependency chain

```
gala-playground
  -> martianoff/gala-server (external module)
       -> martianoff/gala/std (stdlib)
       -> martianoff/gala/collection_immutable (stdlib)
       -> martianoff/gala/concurrent (stdlib)
       -> martianoff/gala-server/httpcore (Go subpackage within gala-server)
```

The build system needs to resolve this full chain and generate correct `require`/`replace` entries for all of them.

## Suggested Fix

1. Read the dependency's `gala.mod` to get its internal module path (e.g., `martianoff/gala-server`)
2. Use that path (not the GitHub URL) in the generated `go.mod`'s `require` and `replace` directives
3. Recursively resolve transitive dependencies and add their `require`/`replace` entries
4. For Go subpackages within the dependency (e.g., `martianoff/gala-server/httpcore`), ensure the parent `replace` covers them
