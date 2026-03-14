# [FIXED in 0.17.0] Build System: Generated go.mod Uses `go 1.21`, Breaking Go 1.22+ Features

## Summary

The GALA build system generates `go 1.21` in the build workspace's `go.mod`. This prevents Go 1.22+ features from working, including the enhanced `http.ServeMux` method-based pattern matching used by `gala-server`.

## Environment

- GALA version: 0.16.1
- Go version: 1.25.5
- OS: Windows 11

## Reproduction

1. Build any project that depends on `gala-server`:
   ```bash
   gala build .
   ```

2. Inspect the generated `go.mod`:
   ```bash
   cat ~/.gala/build/<hash>/go.mod
   ```

   Shows:
   ```
   go 1.21
   ```

3. `gala-server`'s `httpcore/httpcore.go` registers routes using Go 1.22+ method-prefixed patterns:
   ```go
   pattern := r.method + " " + r.pattern  // e.g., "GET /"
   mux.HandleFunc(pattern, handler)
   ```

4. With `go 1.21` in go.mod, Go's `http.ServeMux` treats `"GET /"` as a **literal path**, not a method+path pattern. All requests return 404.

## Error

```
curl http://127.0.0.1:3000/        → "404 page not found"
curl http://127.0.0.1:3000/api/version → "404 page not found"
```

The server starts and logs correctly, but no routes match.

## Fix

Changing `go 1.21` → `go 1.22` in the build workspace's `go.mod` resolves the issue:

```bash
# Manual fix in build workspace:
sed -i 's/^go 1.21$/go 1.22/' ~/.gala/build/<hash>/go.mod
go build ./gen/
# Server now works correctly
```

## Suggested Fix

The GALA build system should generate `go 1.22` (or later) in the build workspace's `go.mod`. Go 1.22 has been stable since February 2024 and is required for:
- Enhanced `http.ServeMux` with method-based routing (`"GET /path"`)
- Per-iteration loop variable semantics (fixing closure capture bugs)
- Other standard library improvements

Since GALA already requires Go 1.21+, bumping to 1.22 should not be a breaking change.
