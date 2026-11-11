FROM golang:1.27 AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN \
  --mount=type=cache,target=/go/pkg/mod \
  go mod download
COPY . .
RUN \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"

FROM debian:13-slim AS final
COPY --from=build /build/slr /usr/local/bin/slr

FROM debian:13-slim AS final_with_ca_certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /build/slr /usr/local/bin/slr

FROM final
