# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o messagarr ./cmd/messagarr

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/messagarr .

# Copy default config
COPY --from=builder /build/config ./config

# Create non-root user
RUN addgroup -g 1000 messagarr && \
  adduser -D -u 1000 -G messagarr messagarr && \
  chown -R messagarr:messagarr /app

USER messagarr

EXPOSE 4545

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4545/health || exit 1

CMD ["./messagarr"]
