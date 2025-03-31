FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ENV GOOS=linux
ENV GOARCH=amd64
RUN CGO_ENABLED=0 go build -o server .

# Create a minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Create config directory
RUN mkdir -p /config

RUN apk add --no-cache tzdata

# Set environment variables
ENV PORT=8080 \
  SUPER_ADMIN_KEY=super_admin_key \
  CONFIG_FILE_PATH=/config/config.json \
  AUTO_SAVE_INTERVAL=60 \
  TZ=Europe/Vienna

# Expose the application port
EXPOSE ${PORT}

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]
