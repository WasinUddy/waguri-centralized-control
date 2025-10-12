# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy the entire workspace to get all dependencies
COPY . .

# Set working directory to the DNS app
WORKDIR /app/apps/dns

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dns main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for SSL/TLS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/apps/dns/dns .

# Copy config file if it exists
COPY --from=builder /app/configs/dns.yaml ./configs/dns.yaml

# Expose DNS port
EXPOSE 53/udp

# Command to run
CMD ["./dns"]
