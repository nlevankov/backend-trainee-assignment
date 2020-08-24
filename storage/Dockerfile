FROM postgres

ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD 123
ENV POSTGRES_DB bta_dev

COPY init_db.sql /docker-entrypoint-initdb.d/
