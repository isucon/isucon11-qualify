### builder01: benchmarker
FROM golang:1.16.5-alpine3.13 AS builder01
WORKDIR /workdir
COPY bench/go.mod bench/go.sum ./bench/
COPY extra/initial-data ./extra/initial-data
WORKDIR /workdir/bench
RUN go mod download
COPY bench/ ./
RUN go build -o bench main.go

### builder02: dockerize
FROM golang:1.16.5-alpine3.13 AS builder02
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

### runner
FROM alpine:3.13 AS runner
WORKDIR /
ARG app app
COPY --from=builder01 /workdir/bench/bench ./
COPY --from=builder02 /usr/local/bin/dockerize /usr/local/bin/
