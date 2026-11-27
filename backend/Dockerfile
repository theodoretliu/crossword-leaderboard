# The builder runs on the build machine's own platform and cross-compiles, so
# deploys from an Apple Silicon laptop don't compile under amd64 emulation.
FROM --platform=$BUILDPLATFORM golang:1.27.4 AS build
ARG TARGETOS TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd cmd
COPY internal internal

# Static binaries (nothing here needs cgo), stripped of debug info.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/server ./cmd/scraper ./cmd/scrapeall

FROM alpine:3.24
# ca-certificates: TLS to Postgres, NYT, and Sentry. curl: the api healthcheck
# in docker-compose.yml.
RUN apk add --no-cache ca-certificates curl
# The .bin names are what docker-compose.yml's entrypoints refer to.
COPY --from=build /out/server /bin/server.bin
COPY --from=build /out/scraper /bin/scraper.bin
COPY --from=build /out/scrapeall /bin/scrapeall.bin
ENTRYPOINT ["/bin/server.bin"]
