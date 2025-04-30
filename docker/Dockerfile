# docker/Dockerfile
FROM golang:1.19-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o bill-aggregator ./cmd/server

# Create the final image
FROM alpine:3.16

# Set the working directory
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/bill-aggregator /app/bill-aggregator

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/bill-aggregator"]
