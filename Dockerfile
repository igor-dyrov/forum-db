FROM ubuntu:18.04

# MAINTAINER Igor Dyrov

# Обвновление списка пакетов
RUN apt-get -y update

#
# Установка postgresql
#
RUN echo 1
ENV PGVER 10
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

USER postgres

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




USER root

RUN cd ~ && mkdir Project3 && cd Project3
# RUN git clone https://github.com/igor-dyrov/forum-db

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq

# Объявлем порт сервера
EXPOSE 5000

# Копируем исходный код в Docker-контейнер

WORKDIR $GOPATH/src/github.com/igor-dyrov/forum-db
ADD . $GOPATH/src/github.com/igor-dyrov/forum-db


USER postgres

CMD service postgresql start && psql -f ./init.sql forum && go run src/main.go
