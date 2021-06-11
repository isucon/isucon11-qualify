FROM golang:1.16.5-alpine3.13

WORKDIR /development
COPY development/backend/air.toml .

WORKDIR /go/src/isucondition

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && go get -u github.com/cosmtrek/air

COPY webapp/go/go.mod .
COPY webapp/go/go.sum .

RUN go mod download