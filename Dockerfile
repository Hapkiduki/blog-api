# ──────────────────────────────────────────────
# Stage 1: Build
# ──────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install git (needed for go mod download) and ca-certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first (Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 produces a static binary (no C dependencies)
# -ldflags="-s -w" strips debug info for a smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server cmd/server/main.go

# ──────────────────────────────────────────────
# Stage 2: Run
# ──────────────────────────────────────────────
FROM alpine:3.19

# Install ca-certificates for HTTPS calls
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the binary from the builder stage
COPY --from=builder /app/server .

# Copy migrations (needed for runtime migration)
COPY --from=builder /app/migrations ./migrations

# Expose the application port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/server"]