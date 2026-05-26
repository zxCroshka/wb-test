FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /app/bin/trending-service \
    ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

RUN adduser -D -u 1000 appuser

WORKDIR /app

RUN mkdir -p /app/data /app/configs

COPY --from=builder /app/bin/trending-service /app/trending-service
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml

RUN chmod +x /app/trending-service

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/trending-service"]
CMD ["-config", "/app/configs/config.yaml"]
