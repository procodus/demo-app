# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o demo-app ./cmd

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /build/demo-app /app/demo-app

# Use non-root user
USER appuser

# Expose ports (documentation only)
# Generator metrics: 9091
# Backend gRPC: 50051, Backend metrics: 9090
# Frontend HTTP: 8080
EXPOSE 9091 50051 9090 8080

# Set entrypoint
ENTRYPOINT ["/app/demo-app"]
