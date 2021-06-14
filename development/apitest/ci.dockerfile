FROM golang:1.16.5-alpine3.13 AS builder

WORKDIR /workdir
COPY bench/go.mod bench/go.sum ./
RUN go mod download
COPY bench/ ./
RUN go build -o bench main.go

FROM alpine:3.13
ARG app app
COPY --from=builder /workdir/bench .
ENTRYPOINT ["./bench", "--no-load"]
