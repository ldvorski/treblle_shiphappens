# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
# git: for go modules
# gcc, musl-dev: for CGO (required by SQLite)
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO is needed for SQLite (modernc.org/sqlite)
# -ldflags="-w -s" strips debug information for smaller binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o api_monitor ./cmd/server/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
# ca-certificates: for HTTPS requests to Jikan API
# tzdata: for timezone support
# wget: for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/api_monitor .

# Create directory for database with proper permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check - verifies the service is responding
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./api_monitor"]

