FROM python:3.9.6-buster

#install mariadb-client
RUN apt-get update \
    && apt-get install -y default-mysql-client

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
ENTRYPOINT ["dockerize", "-wait=tcp://mysql-backend:3306", "-timeout=60s"]

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY webapp/public /webapp/public
COPY webapp/python /webapp/python 

WORKDIR /webapp/python

RUN pip install -r requirements.txt

CMD ["python3", "main.py"]