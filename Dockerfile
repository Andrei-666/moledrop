# Stage 1: build
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy dependency files first so Docker caches this layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -o signaling ./cmd/signaling

# Stage 2: minimal runtime image
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/signaling .

EXPOSE 8080
CMD ["./signaling"]
