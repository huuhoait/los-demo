# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o underwriting-worker ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/underwriting-worker .

# Copy configuration files
COPY --from=builder /app/config ./config

# Create necessary directories
RUN mkdir -p /var/log/underwriting-worker

# Expose port (if needed for health checks)
EXPOSE 8083

# Set environment variables
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD ps aux | grep underwriting-worker | grep -v grep || exit 1

# Run the binary
CMD ["./underwriting-worker"]
