FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o issueforge ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache redis bash

WORKDIR /app

COPY --from=builder /app/issueforge .

EXPOSE 8080

CMD sh -c "redis-server --daemonize yes && until redis-cli ping > /dev/null 2>&1; do sleep 0.1; done; ./issueforge"
