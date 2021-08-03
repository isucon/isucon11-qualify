FROM mariadb:10.3.29-focal

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
COPY mysql-backend/mysql.cnf /etc/mysql/conf.d/mysql.cnf

CMD [ "mysqld" ]
