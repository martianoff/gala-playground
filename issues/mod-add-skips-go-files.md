# [FIXED in 0.15.2] `gala mod add` Skips .go Files When Fetching External Modules

## Summary

When `gala mod add` fetches an external GALA module from GitHub, it only downloads `.gala`, `BUILD.bazel`, and `gala.mod` files. All `.go` files are skipped, including pure Go subpackages that the module depends on internally.

This is the root cause of the `httpcore is not in std` build failure — the file was never fetched in the first place.

## Environment

- GALA version: 0.15.2
- OS: Windows 11

## Reproduction

```bash
gala mod add github.com/martianoff/gala-server@3c657658dcb4
```

Then inspect the downloaded module:

```bash
ls ~/.gala/pkg/mod/github.com/martianoff/gala-server@<hash>/httpcore/
# Only BUILD.bazel — httpcore.go is missing
```

Compare with the actual repo contents:

```bash
gh api repos/martianoff/gala-server/contents/httpcore --jq '.[].name'
# BUILD.bazel
# httpcore.go     <- this file was not fetched
```

## Full list of missing files

Files present in the GitHub repo but missing from the fetched module:

| File | Type | Purpose |
|------|------|---------|
| `httpcore/httpcore.go` | Go source | HTTP bridge (net/http wrapper) |
| `go.mod` | Go module | Go dependency declarations |
| `go.sum` | Go checksums | Go dependency verification |
| `.gitignore` | Config | Git ignore rules |

## Impact

The build system correctly transpiles `.gala` files and copies the output to the build workspace. But since the `.go` source files were never downloaded, the pure Go subpackage `httpcore/` is empty, causing:

```
package martianoff/gala-server/httpcore is not in std
```

## Suggested Fix

`gala mod add` should download **all** files from the module repository, not just `.gala` and build-system files. At minimum, it must include:

1. All `.go` files (required for Go compilation)
2. `go.mod` and `go.sum` (required for Go module resolution)
3. All subdirectories recursively (Go subpackages)

The simplest approach: `git clone` or `git archive` the full repo at the specified commit, then store it in `~/.gala/pkg/mod/`.
