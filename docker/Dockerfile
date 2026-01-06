# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache build-base
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/resources /app/resources
COPY --from=builder /app/migrations /app/migrations
ENV APP_ENV=docker
EXPOSE 8080
ENTRYPOINT ["/app/server"]
