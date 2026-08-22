# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o /out/maestro ./cmd/maestro

FROM alpine:3.22 AS prep
RUN mkdir -p /var/log/maestro && chown 65532:65532 /var/log/maestro

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=prep --chown=65532:65532 /var/log/maestro /var/log/maestro
COPY --from=build /out/maestro /usr/local/bin/maestro

ENV LOGGER_FOLDER=/var/log/maestro \
    SERVER_ADDR=:8080

EXPOSE 8080

USER 65532

ENTRYPOINT ["/usr/local/bin/maestro"]
