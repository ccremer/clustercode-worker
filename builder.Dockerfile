#______________________________________________________________________________
#### Base Image, to save build time on local dev machine
FROM golang:1.11-alpine

WORKDIR /go/src/app

COPY ["go.mod", "./"]

RUN \
    apk add --no-cache git build-base && \
    env GO111MODULE=on go mod download
