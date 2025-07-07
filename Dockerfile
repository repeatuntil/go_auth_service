FROM golang:1.24 as builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o app .

FROM alpine:latest as runner

WORKDIR /auth_service

COPY --from=builder /build/database/migrations ./database/migrations
COPY --from=builder /build/app .
COPY --from=builder /build/.env .

EXPOSE 8080

CMD ["./app"]