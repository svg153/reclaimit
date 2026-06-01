# syntax=docker/dockerfile:1

# ---- Build stage ---------------------------------------------------------
FROM golang:1.24-alpine AS builder

# Build metadata injected at build time (override with --build-arg).
ARG VERSION=dev

WORKDIR /src

# Cache dependencies separately from the source for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static, stripped, reproducible binary. CGO is disabled so the result runs on
# scratch/distroless without libc.
RUN CGO_ENABLED=0 GOOS=linux \
	go build -trimpath -ldflags "-s -w -X github.com/svg153/reclaimit.Version=${VERSION}" \
	-o /out/reclaimit ./cmd/reclaimit

# ---- Runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="reclaimit" \
	org.opencontainers.image.description="Reclaim disk space by finding and removing regenerable build/cache directories." \
	org.opencontainers.image.source="https://github.com/svg153/reclaimit" \
	org.opencontainers.image.licenses="MIT"

COPY --from=builder /out/reclaimit /usr/local/bin/reclaimit

# Run as the non-root user provided by the distroless image.
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/reclaimit"]
CMD ["--help"]
