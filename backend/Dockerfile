FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -o analyze ./cmd/analyze

FROM python:3.12-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
        tzdata ffmpeg ca-certificates && \
    pip install --no-cache-dir yt-dlp faster-whisper && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/analyze .
COPY migrations/ ./migrations/
COPY scripts/ ./scripts/
EXPOSE 8080
CMD ["./server"]
