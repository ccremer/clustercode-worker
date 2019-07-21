#______________________________________________________________________________
#### Base Image, to save build time on local dev machine
ARG GOARCH
ARG ARCH
FROM golang:1.12-alpine as builder

WORKDIR /go/src/app

COPY ["go.mod", "./"]

RUN \
    apk add --no-cache git build-base && \
    env GO111MODULE=on go mod download

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified

COPY / .
RUN \
    pwd && \
    env GO111MODULE=on go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT}"

#______________________________________________________________________________
#### Runtime Image
ARG ARCH
FROM multiarch/alpine:${ARCH} as runtime

RUN \
    apk add --no-cache curl ffmpeg bash && \
    # Let's create the directories first so we can apply the permissions:
    mkdir -m 755 /usr/share/clustercode && \
    mkdir -m 777 /input /output /var/tmp/clustercode

VOLUME \
    /input \
    /output \
    /var/tmp/clustercode

ENTRYPOINT ["clustercode-worker"]
CMD ["-c", "clustercode"]

COPY --from=builder /go/src/app/clustercode-worker /usr/bin/clustercode-worker
RUN \
    clustercode-worker --save-config /usr/share/clustercode/clustercode.yaml
USER 1000
