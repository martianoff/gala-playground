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

ARG GALA_VERSION=0.32.0
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

# Create non-root user BEFORE pre-warming so caches are owned by the right user
RUN adduser -D -h /home/gala gala && \
    chown -R gala:gala /go
USER gala
WORKDIR /home/gala

# Copy examples for pre-warming
COPY --from=builder /build/examples/ /tmp/examples/

# Pre-warm: build every predefined example to populate analysis cache and
# Go build cache for all import combinations. Uses the same workspace path
# the server will use at runtime (/tmp/gala-playground-ws).
# Runs as 'gala' user — caches land in /home/gala/.gala/ and /home/gala/.cache/
RUN mkdir -p /tmp/gala-playground-ws && \
    printf 'module playground\n\ngala 0.32.0\n' > /tmp/gala-playground-ws/gala.mod && \
    for example in /tmp/examples/*.gala; do \
        name=$(basename "$example" .gala); \
        cp "$example" /tmp/gala-playground-ws/main.gala; \
        if gala build -o /tmp/gala-playground-ws/bin /tmp/gala-playground-ws 2>/dev/null; then \
            echo "  warmed: $name"; \
        else \
            echo "  skip:   $name (build error)"; \
        fi; \
    done && \
    rm -rf /tmp/examples 2>/dev/null; \
    echo "All examples warmed"

EXPOSE 3000

# Bind to all interfaces inside container
ENV BIND_ALL=1

ENTRYPOINT ["playground"]
