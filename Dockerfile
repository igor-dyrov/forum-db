FROM ubuntu:16.04

MAINTAINER Igor Dyrov

# Обвновление списка пакетов
RUN apt-get -y update

#
# Установка postgresql
#
RUN echo 1
ENV PGVER 9.6
RUN apt-get install -y postgresql-$PGVER

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker forum &&\
    /etc/init.d/postgresql stop

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

#
# Сборка проекта
#

# Установка golang
RUN apt install -y golang-1.10 git

# Выставляем переменную окружения для сборки проекта
ENV GOROOT /usr/lib/go-1.10
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

# Копируем исходный код в Docker-контейнер
#WORKDIR $GOPATH/src/github.com/igor-dyrov/forum-db/src
#ADD golang/ $GOPATH/src/github.com/bozaro/igor-dyrov/forum-db/src
#ADD common/ $GOPATH/src/github.com/bozaro/igor-dyrov/forum-db/src

## Собираем генераторы
#RUN go install ./vendor/github.com/go-swagger/go-swagger/cmd/swagger
#RUN go install ./vendor/github.com/jteeuwen/go-bindata/go-bindata
#
## Собираем и устанавливаем пакет
#RUN go generate -x ./restapi
#RUN go install ./cmd/hello-server

USER root

RUN cd ~ && mkdir Project8 && cd Project8
RUN git clone https://github.com/igor-dyrov/forum-db

USER postgres

RUN /etc/init.d/postgresql start && psql -f ./forum-db/init.sql forum && /etc/init.d/postgresql stop

USER root

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq

# Объявлем порт сервера
EXPOSE 5000


#
# Запускаем PostgreSQL и сервер
#
CMD service postgresql start && go run forum-db/src/main.go
