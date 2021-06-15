FROM golang:1.16.5-buster

WORKDIR /development
COPY development/backend-go/air.toml .

#install mysql-client
RUN wget https://dev.mysql.com/get/mysql-apt-config_0.8.17-1_all.deb \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y ./mysql-apt-config_0.8.17-1_all.deb \
    && apt-get update \
    && apt-get install -y mysql-client  \
    && rm ./mysql-apt-config_0.8.17-1_all.deb
    
WORKDIR /go/src/isucondition

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && go get -u github.com/cosmtrek/air

COPY webapp/go/go.mod .
COPY webapp/go/go.sum .

RUN go mod download
