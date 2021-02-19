FROM ubuntu:focal

RUN apt update && apt install mysql-client postgresql-client-common postgresql-client mongodb-clients -y