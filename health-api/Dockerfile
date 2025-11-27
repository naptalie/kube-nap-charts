# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version info
ARG BUILD_REF=develop
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.build=${BUILD_REF}" -a -installsuffix cgo -o health-api ./app/services/health-api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/health-api .

# Expose API and debug ports
EXPOSE 8080 4000

# Run the application
CMD ["./health-api"]
