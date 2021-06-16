FROM golang:1.16.5-alpine3.13

WORKDIR /webapp/mysql/db
COPY webapp/mysql/db/ .

WORKDIR /webapp/go

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

COPY webapp/go/go.mod webapp/go/go.sum ./
RUN go mod download

COPY webapp/go/ .
RUN go build -o app .

ENTRYPOINT ["dockerize", "-wait=tcp://mysql-backend:3306", "-timeout=60s", "./app"]
