FROM golang:1.23.8-alpine AS builder

WORKDIR /app

COPY ../go.mod go.sum ./
COPY ../cmd ./cmd
COPY ../internal ./internal
COPY ../gen ./gen
COPY ../proto ./proto

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o http ./cmd/http

FROM alpine:3.21

WORKDIR /root/

COPY --from=builder /app/http .
COPY ../api/swagger.json /app/http/swagger.json

EXPOSE 8080

CMD ["./http"]