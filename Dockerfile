# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Copy go mod files first (untuk better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application untuk AMD64 (PENTING!)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user untuk security
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Change ownership
RUN chown appuser:appuser main

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 6005

# Set environment variables
ENV GIN_MODE=release
ENV TZ=Asia/Jakarta

# Run the application
CMD ["./main"]