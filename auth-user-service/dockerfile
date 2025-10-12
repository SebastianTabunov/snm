# Build stage - используем Go 1.25
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

# Build application
RUN go build -o main ./cmd/server

# Runtime stage
FROM alpine:latest

# Install postgres client
RUN apk --no-cache add postgresql-client

WORKDIR /app

# Copy binaries and files
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

EXPOSE 8080

# Run migrations and start app
CMD ["sh", "-c", "/app/scripts/migrate.sh && /app/main"]
