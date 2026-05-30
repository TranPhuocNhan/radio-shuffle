# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine AS deps

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM deps AS build

ARG TARGET=api
ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    case "$TARGET" in api|syncer) ;; *) echo "unsupported TARGET=$TARGET" >&2; exit 1 ;; esac && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -buildid=" -o /out/app ./cmd/$TARGET

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=build /out/app /app/app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/app"]
