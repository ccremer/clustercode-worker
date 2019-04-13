FROM golang:1.11-alpine

WORKDIR /go/src/app

COPY ["go.mod", "./"]

ENV \
    GO111MODULE=on

RUN \
    apk add --no-cache git build-base libxml2-dev && \
    go mod download

COPY / .

RUN \
    go build && \
    ./clustercode-worker
