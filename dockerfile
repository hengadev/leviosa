# TODO: Learn how to setup the env variable using the .env file
# TODO: Handle the frontend part of the application
FROM golang:1.21.5-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin .


# FROM alpine:edge
FROM scratch

WORKDIR /app

ENV PORT=3000

COPY --from=builder /app/bin .

ENTRYPOINT ["/app/bin"]
