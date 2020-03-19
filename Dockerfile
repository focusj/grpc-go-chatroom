FROM golang:latest

ENV GRPC_GO_LOG_SEVERITY_LEVEL=info \
    GRPC_GO_LOG_VERBOSITY_LEVEL=2

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN go build -o main .

CMD ["/app/main"]