# base.dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CMD_PATH
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/${CMD_PATH}/main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/server .

CMD ["./server"]