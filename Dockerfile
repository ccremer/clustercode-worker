#______________________________________________________________________________
#### Base Image, to save build time on local dev machine
ARG GOARCH=amd64
ARG ARCH=amd64-edge
ARG ALPINE_REPO=http://dl-3.alpinelinux.org/alpine/edge/main/
FROM docker.io/library/golang:1.13-alpine as builder

WORKDIR /go/src/app

COPY ["go.mod", "./"]

RUN \
    apk add git build-base --no-cache --repository "${ALPINE_REPO}" && \
    env GO111MODULE=on go mod download

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified

COPY / .
RUN \
    env GO111MODULE=on go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT}"

#______________________________________________________________________________
#### Runtime Image
ARG ARCH=amd64-edge
ARG ALPINE_REPO=http://dl-3.alpinelinux.org/alpine/edge/main/
FROM docker.io/multiarch/alpine:${ARCH} as runtime

ENTRYPOINT ["clustercode-worker"]
EXPOSE 8080

RUN \
    apk add curl ffmpeg bash --no-cache --repository "${ALPINE_REPO}" && \
    # Let's create the directories first so we can apply the permissions:
    mkdir -m 755 /usr/share/clustercode && \
    mkdir -m 777 /input /output /var/tmp/clustercode

VOLUME \
    /input \
    /output \
    /var/tmp/clustercode

COPY --from=builder /go/src/app/clustercode-worker /usr/bin/clustercode-worker
USER 1000:0
