FROM perl:5.34.0-buster

RUN apt-get update && apt-get install -y wget default-mysql-client

WORKDIR /usr/local/bin

RUN curl -fsSL --compressed https://raw.githubusercontent.com/skaji/cpm/master/cpm > cpm \
    && chmod +x cpm

WORKDIR /webapp/perl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY webapp/perl/cpanfile* .

RUN cpm install -g --show-build-log-on-failure

COPY webapp/perl .
COPY webapp/public /webapp/public

ENTRYPOINT ["dockerize", "-wait=tcp://mysql-backend:3306", "-timeout=60s"]

CMD ["plackup", "-p", "3000", "-Ilib", "app.psgi"]
