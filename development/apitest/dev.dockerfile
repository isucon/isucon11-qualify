FROM golang:1.16.5-buster AS runner
RUN apt-get update && apt-get install -y gcc g++

WORKDIR /workdir
COPY bench/go.mod bench/go.sum ./
RUN go mod download

