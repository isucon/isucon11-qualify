# build stage
FROM golang:1.16.5 as builder
## init setting
WORKDIR /extra/jiaapi-mock/
## download packages
COPY extra/jiaapi-mock/go.mod extra/jiaapi-mock/go.sum ./
COPY bench/random /bench/random
RUN go mod download
## build
COPY extra/jiaapi-mock/ ./
RUN GOOS=linux make

# run stage
FROM gcr.io/distroless/base
## copy binary
COPY --from=builder /extra/jiaapi-mock/jiaapi-mock .
## Run
ENTRYPOINT ["./jiaapi-mock"]

