# BUILD STAGE
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bot-app ./cmd/api

# RUN STAGE
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bot-app .
CMD ["./bot-app"]