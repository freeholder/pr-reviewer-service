FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pr-service ./cmd/pr-service

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/pr-service ./pr-service
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENV APP_HTTP_PORT=8080

CMD ["./pr-service"]
