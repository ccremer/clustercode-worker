#______________________________________________________________________________
#### Builder Image
ARG GOARCH
ARG ARCH
FROM braindoctor/clustercode-worker:builder as builder

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified

COPY / .
RUN \
    pwd && \
    env GO111MODULE=on go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT}"

#______________________________________________________________________________
#### Runtime Image
ARG ARCH
FROM multiarch/alpine:${ARCH}

RUN \
    apk add --no-cache curl ffmpeg libxml2 && \
    # Let's create the directories first so we can apply the permissions:
    mkdir -m 755 /usr/share/clustercode && \
    mkdir -m 777 /input /output /var/tmp/clustercode

VOLUME \
    /input \
    /output \
    /var/tmp/clustercode

ENTRYPOINT ["worker"]

COPY schema/clustercode_v1.xsd /usr/share/clustercode/
COPY --from=builder /go/src/app/clustercode-worker /usr/bin/worker
USER 1000
