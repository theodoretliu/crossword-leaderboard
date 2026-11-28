# The builder runs on the build machine's own platform and cross-compiles, so
# deploys from an Apple Silicon laptop don't compile under amd64 emulation.
FROM --platform=$BUILDPLATFORM golang:1.27.4 AS build
ARG TARGETOS TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd cmd
COPY internal internal

# One static binary (nothing here needs cgo), stripped of debug info.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/crossword ./cmd/crossword

FROM alpine:3.24
# ca-certificates: TLS to Postgres, NYT, and Sentry. curl: the api healthcheck
# in docker-compose.yml.
RUN apk add --no-cache ca-certificates curl
COPY --from=build /out/crossword /bin/crossword
# The binary picks its command from the name it's run as. These .bin names are
# what docker-compose.yml's entrypoints (and images from before the merge) use,
# so rolling back to an older image tag keeps working.
RUN ln -s crossword /bin/server.bin && \
    ln -s crossword /bin/scraper.bin && \
    ln -s crossword /bin/scrapeall.bin
ENTRYPOINT ["/bin/server.bin"]
