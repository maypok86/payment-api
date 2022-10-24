FROM alpine:latest

ENV GOOSE_FILE /bin/goose
ENV MIGRATIONS /migrations

ARG GOOSE_VERSION
# Add goose binary file -------------------------
ADD https://github.com/pressly/goose/releases/download/v${GOOSE_VERSION}/goose_linux_x86_64 ${GOOSE_FILE}
RUN chmod +x ${GOOSE_FILE}

WORKDIR ${MIGRATIONS}

CMD ${GOOSE_FILE} postgres "host=postgres user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DBNAME} sslmode=${POSTGRES_SSLMODE}" up