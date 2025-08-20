# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    openssh-client \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 chisel && \
    adduser -D -u 1001 -G chisel chisel

# Copy binary from builder stage
COPY --from=builder /app/bin/chisel /usr/local/bin/chisel

# Set ownership and permissions
RUN chown chisel:chisel /usr/local/bin/chisel && \
    chmod +x /usr/local/bin/chisel

# Switch to non-root user
USER chisel

# Set working directory
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["chisel"]
CMD ["--help"]
