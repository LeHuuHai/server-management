FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy all go workspace files
COPY . .

# Download dependencies
RUN go mod download

# Build mail-worker
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./mail-worker/cmd/main.go

FROM alpine:3.19

# Install CA certificates for SMTP TLS
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]
