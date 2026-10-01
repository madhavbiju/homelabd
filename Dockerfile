FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a statically linked binary
# CGO_ENABLED=0 is required for a truly static binary, though modernc.org/sqlite is pure Go anyway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o homelabd ./cmd/homelabd

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/homelabd /app/homelabd

# Create a non-root user (optional, but good practice). 
# However, if mounting /var/run/docker.sock, the user typically needs to be in the 'docker' group or root.
# For simplicity in homelab environments, we default to root, but it can be overridden.

EXPOSE 8080

ENTRYPOINT ["/app/homelabd"]
