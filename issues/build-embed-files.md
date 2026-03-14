# [FIXED in 0.15.0] Embedded Files Not Copied to Build Workspace

## Summary

When using `embed val` directives, the referenced files/directories are not copied to the GALA build workspace (`~/.gala/build/<hash>/gen/`). The Go compiler then fails with `pattern <glob>: no matching files found`.

## Environment

- GALA version: 0.14.0
- OS: Windows 11

## Reproduction

Project structure:

```
myproject/
  main.gala
  gala.mod
  static/
    index.html
  examples/
    hello.gala
```

`main.gala`:

```gala
package main

embed val indexHTML = "static/index.html"
embed val exampleFiles EmbeddedFS = "examples/*"

func main() {
    Println(indexHTML)
}
```

Run `gala build .`

## Expected

The build system copies `static/` and `examples/` directories into the build workspace alongside the generated `.gen.go` files, so Go's embed directives can resolve them.

## Actual

```
gen\main.gen.go:23:12: pattern examples/*: no matching files found
```

The generated Go code contains correct `//go:embed` directives:

```go
//go:embed static/index.html
var indexHTML string
//go:embed examples/*
var _embed_exampleFiles embed.FS
```

But `static/` and `examples/` don't exist in the build workspace at `~/.gala/build/<hash>/gen/`.

## Suggested Fix

During `gala build`, scan for `embed val` declarations and copy the referenced files/directories from the source project into the build workspace's `gen/` directory.
