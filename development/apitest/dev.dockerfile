FROM golang:1.16.5-buster AS runner
RUN apt-get update && apt-get install -y gcc g++

#install ca.crt
COPY development/certificates/ca.crt /usr/local/share/ca-certificates/extra/ca.crt
RUN apt-get update -y && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/* && update-ca-certificates

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

WORKDIR /workdir
COPY extra/initial-data /extra/initial-data
COPY bench/go.mod bench/go.sum ./
RUN go mod download

