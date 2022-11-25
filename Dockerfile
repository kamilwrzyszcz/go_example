### Multistage
# Build stage
FROM golang:1.19.2-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
#
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait wait
RUN chmod +x wait
#

# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY --from=builder /app/wait .
COPY app.env .
COPY start.sh .
COPY db/migration ./migration

EXPOSE 8080