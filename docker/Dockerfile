# Stage 1: Download all dependencies
FROM golang:1.16.4-alpine

WORKDIR /usr/local/go/src/github.com/opentibiabr/login-server
COPY go.mod .
COPY go.sum .
COPY src ./src

RUN go build -o /bin/login-server ./src/

VOLUME /login-server

ENTRYPOINT ["/bin/login-server"]

