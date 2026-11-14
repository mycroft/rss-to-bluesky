# syntax=docker/dockerfile:1

# The build stage runs on the build platform and cross-compiles for the target.
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27 AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=devel
ARG COMMIT=""

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/rss-to-bluesky ./cmd/rss-to-bluesky && \
    mkdir -p /out/database

# Static binary, no shell: distroless ships only CA certificates and tzdata.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/rss-to-bluesky /usr/local/bin/rss-to-bluesky
# The state database is ./database/db, relative to the working directory.
# Mount a volume on /app/database to keep it across runs: it holds the GUIDs
# already posted, and losing it reposts the whole feed.
WORKDIR /app
COPY --from=build --chown=65532:65532 /out/database /app/database
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/rss-to-bluesky"]
