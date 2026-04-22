# GALA Playground

[![GALA](https://img.shields.io/badge/GALA-0.32.0-blue)](https://github.com/martianoff/gala)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](#docker)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A web-based playground for the [GALA programming language](https://github.com/martianoff/gala), inspired by [Go Playground](https://go.dev/play/). Write and run GALA code directly in your browser.

**Try it live: [gala-playground.fly.dev](https://gala-playground.fly.dev)**

---

## Features

- **Browser-based code editor** with line numbers, tab support, and resizable split panels
- **Instant execution** -- compile and run GALA code with one click or `Ctrl+Enter`
- **15 built-in examples** covering all major GALA features
- **Error display** -- compile errors and runtime panics shown inline
- **Share** -- copy code to clipboard with one click
- **Dark theme** -- developer-friendly IDE-like interface

## Built-in Examples

| Example | Features Showcased |
|---------|--------------------|
| Hello | Basic syntax, `Println`, `func main()` |
| Sealed Types | ADTs, exhaustive pattern matching, expression functions |
| Pattern Matching | Guards, custom extractors (`Unapply`), type matching |
| Collections | `Array`, `Filter`, `Map`, `FoldLeft`, functional pipelines |
| Option & Either | `Some`/`None`, `Left`/`Right`, safe null handling, validation |
| Try Monad | Panic recovery, `Map` chaining, `Recover`, `GetOrElse` |
| HashMap | Immutable maps, `FoldLeft` word counter, `Sorted`, `MapValues` |
| Generics | Type parameters, generic structs, generic methods |
| Roman to Int | String processing, `FoldLeft`, `match` on characters |
| Streams | Lazy evaluation, infinite sequences, `Take`, `Filter`, `Map` |
| Named Arguments | Default params, named args in any order, `Copy` overrides |
| JSON | `JsonStringify`, `JsonParse`, `Json[T]` extractor, `Collect` |
| Regex | Capture groups via `Unapply` + `Array` destructuring, `FindAll` |
| IO Effects | `IO[T]` monad, `Suspend`, `Map`/`FlatMap`, `Recover`, `AndThen` |
| Tour | Full 10-section showcase with formatted output |

---

## Quick Start

### Local

Requires [Go 1.25+](https://go.dev/dl/) and [GALA 0.32.0+](https://github.com/martianoff/gala/releases).

```bash
# Install GALA
curl -fsSL -o ~/.local/bin/gala \
  https://github.com/martianoff/gala/releases/latest/download/gala-linux-amd64
chmod +x ~/.local/bin/gala

# Clone and run
git clone https://github.com/martianoff/gala-playground.git
cd gala-playground
gala run .
```

Opens automatically at [http://localhost:3000](http://localhost:3000).

### Docker (pre-built)

```bash
docker run -p 3000:3000 maxmtmn/gala-playground
```

Open [http://localhost:3000](http://localhost:3000).

### Docker (build from source)

```bash
docker build -t gala-playground .
docker run -p 3000:3000 gala-playground
```

---

## Architecture

```
gala-playground/
  main.gala          # GALA server (uses gala-server module)
  gala.mod           # GALA module file with dependencies
  static/
    index.html       # Single-page frontend (editor + output)
  examples/
    hello.gala       # 15 pre-loaded GALA examples
    sealed_types.gala
    ...
  Dockerfile         # Multi-stage build (GALA build + runtime)
```

The server is written entirely in GALA using the [`gala-server`](https://github.com/martianoff/gala-server) HTTP framework. Static files and examples are embedded at compile time using GALA's `embed val` directives.

### How It Works

1. User writes GALA code in the browser editor
2. Clicking **Run** (or `Ctrl+Enter`) sends the code to `/api/run`
3. The server writes the code to a temp directory with a `gala.mod`
4. `gala run` transpiles the GALA code to Go and executes it
5. Output (or errors) are returned and displayed in the output panel

### API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/run` | POST | Execute GALA code. Body: `{"code": "..."}`. Returns: `{"output", "error", "time"}` |
| `/api/examples` | GET | List all built-in examples. Returns: `[{"name", "code"}]` |
| `/api/version` | GET | Returns the GALA version. Returns: `{"version": "0.10.0"}` |
| `/` | GET | Serves the playground UI |

### Limits

- **Code size**: 50 KB max
- **Execution timeout**: 30 seconds
- **Single file**: each run is a single `main.gala` with `package main`

---

## Docker Details

The Dockerfile uses a multi-stage build:

| Stage | Base Image | Purpose |
|-------|------------|---------|
| `builder` | `golang:1.25.5-alpine` | Downloads GALA and compiles `main.gala` via `gala build` |
| Runtime | `golang:1.25.5-alpine` | Runs both the server and `gala run` |

The runtime image includes Go because `gala run` invokes `go build` internally. The GALA stdlib and Go module cache are pre-warmed during the Docker build so the first user request is fast.

### Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `BIND_ALL` | `1` (in Docker) | Bind to `0.0.0.0` instead of `127.0.0.1` |
| `PORT` | `3000` | Server port |

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
