FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache postgresql-client git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Build application - ИСПРАВЛЕНО для cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o auth-service ./cmd/server

# Runtime stage
FROM alpine:latest

# Install postgres client and ca-certificates
RUN apk --no-cache add postgresql-client ca-certificates wget

WORKDIR /app

# Copy binaries and files
COPY --from=builder /app/auth-service .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Make scripts executable
RUN chmod +x /app/scripts/migrate.sh

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run migrations and start app
CMD ["sh", "-c", "/app/scripts/migrate.sh && /app/auth-service"]