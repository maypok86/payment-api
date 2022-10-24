FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "$LDFLAGS" \
    -v -o ./bin/ ./cmd/api

FROM alpine:latest
LABEL maintainer="Mayshev Alex <alex.mayshev.86@gmail.com>"

WORKDIR /app/

COPY --from=builder /app/bin/api .

EXPOSE 8080

CMD ["./api"]