# Stage 1: Download all dependencies
FROM golang:1.16.4-alpine as builder

WORKDIR /usr/local/go/src/github.com/opentibiabr/login-server
COPY go.mod .
COPY go.sum .
COPY src ./src

RUN go build -o /bin/login-server ./src/

FROM alpine:latest

# Copy our static executable.
COPY --from=builder /bin/login-server /bin/login-server

ENTRYPOINT ["/bin/login-server"]

