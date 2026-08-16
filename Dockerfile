FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o issueforge ./cmd/api

FROM alpine:3.22

COPY --from=builder /app/issueforge .

EXPOSE 8080

CMD [ "./issueforge" ]