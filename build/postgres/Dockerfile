# Установка postgresql
FROM ubuntu:18.04 AS release


ENV PGVER 10
RUN apt -y update && apt install -y postgresql-$PGVER
USER postgres
COPY . .
RUN cat ./postgresql.conf >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker myService &&\
    PGPASSWORD=docker psql -U docker -h 127.0.0.1 -d myService -f ./init.sql &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]


COPY dockerPing.sh .

CMD service postgresql start && service postgresql status &&  bash ./dockerPing.sh
