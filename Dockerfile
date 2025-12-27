# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage for server
FROM alpine:latest AS server

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy server binary from builder
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
