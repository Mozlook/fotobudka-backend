FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.23 AS runtime

RUN apk add --no-cache ca-certificates tzdata wget

RUN addgroup -S app && adduser -S -G app app && mkdir -p /app /logs
WORKDIR /app

COPY --from=builder --chown=app:app /out/api /app/api
COPY --from=builder --chown=app:app /out/worker /app/worker

USER app

ENV SIEM_LOG_DIR=/logs

FROM runtime AS api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM runtime AS worker
ENTRYPOINT ["/app/worker"]
