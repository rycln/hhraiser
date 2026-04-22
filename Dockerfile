# syntax=docker/dockerfile:1
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /out/hhraiser \
    ./cmd/hhraiser

FROM alpine:3.21

RUN apk add --no-cache \
    tzdata \
    ca-certificates \
  && addgroup -S hhraiser \
  && adduser -S -G hhraiser hhraiser

COPY --from=builder /out/hhraiser /app/hhraiser

RUN mkdir -p /config && chown hhraiser:hhraiser /config

VOLUME ["/config"]

ENV HH_CONFIG_DIR=/config \
    TZ=UTC \
    LOG_LEVEL=info

USER hhraiser

ENTRYPOINT ["/app/hhraiser"]