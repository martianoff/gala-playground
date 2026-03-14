# ============================================================================
#  GALA Playground — Docker Image
# ============================================================================
#
#  Build:   docker build -t gala-playground .
#  Run:     docker run -p 3000:3000 gala-playground
#  Open:    http://localhost:3000
#
# ============================================================================

# --- Stage 1: Download GALA binary ---
FROM alpine:3.21 AS gala-download

ARG GALA_VERSION=0.17.1
ARG TARGETARCH=amd64

RUN apk add --no-cache curl && \
    curl -fsSL -o /gala \
    "https://github.com/martianoff/gala/releases/download/${GALA_VERSION}/gala-linux-${TARGETARCH}" && \
    chmod +x /gala

# --- Stage 2: Build the playground server with GALA ---
FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache git

COPY --from=gala-download /gala /usr/local/bin/gala

WORKDIR /build

COPY gala.mod gala.sum* go.mod go.sum* ./
COPY main.gala ./
COPY static/ ./static/
COPY examples/ ./examples/

RUN gala mod tidy && gala build -o playground .

# --- Stage 3: Runtime image ---
FROM golang:1.25.5-alpine

# Go is needed at runtime because 'gala run' invokes 'go build' internally
RUN apk add --no-cache git ca-certificates

# Install GALA binary
COPY --from=gala-download /gala /usr/local/bin/gala

# Install playground server
COPY --from=builder /build/playground /usr/local/bin/playground

# Pre-warm: extract GALA stdlib and download Go modules so first request is fast
RUN mkdir -p /tmp/warmup && \
    printf 'module warmup\n\ngala 0.17.1\n' > /tmp/warmup/gala.mod && \
    printf 'package main\n\nimport (\n    "fmt"\n    . "martianoff/gala/collection_immutable"\n)\n\nfunc main() {\n    fmt.Println(ArrayOf(1, 2, 3))\n}\n' > /tmp/warmup/main.gala && \
    gala run /tmp/warmup && \
    rm -rf /tmp/warmup

EXPOSE 3000

# Bind to all interfaces inside container
ENV BIND_ALL=1

# Run as non-root
RUN adduser -D -h /home/gala gala
USER gala
WORKDIR /home/gala

ENTRYPOINT ["playground"]
