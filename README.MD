# GALA Playground

[![GALA](https://img.shields.io/badge/GALA-0.10.0-blue)](https://github.com/martianoff/gala)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](#docker)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A web-based playground for the [GALA programming language](https://github.com/martianoff/gala), inspired by [Go Playground](https://go.dev/play/). Write and run GALA code directly in your browser.

**Try it live: [gala-playground.fly.dev](https://gala-playground.fly.dev)**

---

## Features

- **Browser-based code editor** with line numbers, tab support, and resizable split panels
- **Instant execution** -- compile and run GALA code with one click or `Ctrl+Enter`
- **9 built-in examples** covering all major GALA features
- **Error display** -- compile errors and runtime panics shown inline
- **Share** -- copy code to clipboard with one click
- **Dark theme** -- developer-friendly IDE-like interface

## Built-in Examples

| Example | Features Showcased |
|---------|--------------------|
| Hello | Basic syntax, `fmt`, `func main()` |
| Sealed Types | ADTs, exhaustive pattern matching, expression functions |
| Pattern Matching | Guards, custom extractors (`Unapply`), type matching |
| Collections | `Array`, `Filter`, `Map`, `FoldLeft`, functional pipelines |
| Option & Either | `Some`/`None`, `Left`/`Right`, safe null handling, validation |
| Try Monad | Panic recovery, `Map` chaining, `Recover`, `GetOrElse` |
| HashMap | Immutable maps, `FoldLeft` word counter, `Sorted`, `MapValues` |
| Generics | Type parameters, generic structs, generic methods |
| Tour | Full 10-section showcase with formatted output |

---

## Quick Start

### Local

Requires [Go 1.25+](https://go.dev/dl/) and the [GALA binary](https://github.com/martianoff/gala/releases).

```bash
# Install GALA
curl -fsSL -o ~/.local/bin/gala \
  https://github.com/martianoff/gala/releases/latest/download/gala-linux-amd64
chmod +x ~/.local/bin/gala

# Clone and run
git clone https://github.com/martianoff/gala-playground.git
cd gala-playground
go run server.go
```

Opens automatically at [http://localhost:3000](http://localhost:3000).

### Docker

```bash
docker build -t gala-playground .
docker run -p 3000:3000 gala-playground
```

Open [http://localhost:3000](http://localhost:3000).

---

## Architecture

```
gala-playground/
  server.go          # Go web server (embeds static files + examples)
  static/
    index.html       # Single-page frontend (editor + output)
  examples/
    hello.gala       # 9 pre-loaded GALA examples
    sealed_types.gala
    ...
  Dockerfile         # Multi-stage build (server + GALA binary)
```

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
| `/` | GET | Serves the playground UI |

### Limits

- **Code size**: 50 KB max
- **Execution timeout**: 30 seconds
- **Single file**: each run is a single `main.gala` with `package main`

---

## Docker Details

The Dockerfile uses a 3-stage multi-stage build:

| Stage | Base Image | Purpose |
|-------|------------|---------|
| `builder` | `golang:1.25.5-alpine` | Compiles the playground Go server |
| `gala-download` | `alpine:3.21` | Downloads GALA binary from GitHub releases |
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
