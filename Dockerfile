#______________________________________________________________________________
#### Base Image, to save build time on local dev machine
ARG GOARCH=amd64
ARG ARCH=amd64-edge
FROM docker.io/library/golang:1.13-alpine as builder

WORKDIR /go/src/app

COPY ["go.mod", "./"]

RUN \
    env GO111MODULE=on go mod download

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified

COPY / .
RUN \
    env GO111MODULE=on go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT}"

#______________________________________________________________________________
#### Runtime Image
ARG ARCH=amd64-edge
FROM docker.io/multiarch/alpine:${ARCH} as runtime

ENV ALPINE_MIRROR=http://dl-cdn.alpinelinux.org/alpine
ENTRYPOINT ["clustercode-worker"]
EXPOSE 8080

RUN \
    echo "${ALPINE_MIRROR}/${ALPINE_REL}/main" > /etc/apk/repositories && \
    echo "${ALPINE_MIRROR}/${ALPINE_REL}/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache curl ffmpeg bash && \
    # Let's create the directories first so we can apply the permissions:
    mkdir -m 755 /usr/share/clustercode && \
    mkdir -m 777 /input /output /var/tmp/clustercode

VOLUME \
    /input \
    /output \
    /var/tmp/clustercode

COPY --from=builder /go/src/app/clustercode-worker /usr/bin/clustercode-worker
USER 1000:0
