# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY app/ ./

RUN go build -o manager ./cmd/manager

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/manager /usr/local/bin/manager

CMD ["tail", "-f", "/dev/null"]