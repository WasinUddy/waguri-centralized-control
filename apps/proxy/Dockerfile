# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy the entire workspace to get all dependencies
COPY . .

# Set working directory to the proxy app
WORKDIR /app/apps/proxy

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for SSL/TLS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/apps/proxy/proxy .

# Copy static files
COPY --from=builder /app/apps/proxy/menu.html .

# Copy config file if it exists
COPY --from=builder /app/configs/proxy.yaml ./configs/proxy.yaml

# Expose proxy port
EXPOSE 8080

# Command to run
CMD ["./proxy"]
