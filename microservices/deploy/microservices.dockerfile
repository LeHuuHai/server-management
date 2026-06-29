FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy all go workspace files
COPY . .

# Download dependencies
RUN go mod download

# Accept the service name to build (e.g., auth-service, server-service)
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./${SERVICE_NAME}/cmd/main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]
