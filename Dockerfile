#______________________________________________________________________________
#### Builder Image
ARG GOARCH
ARG ARCH
FROM braindoctor/clustercode-worker:builder as builder

COPY / .
RUN \
    pwd && \
    env GO111MODULE=on go build

#______________________________________________________________________________
#### Runtime Image
ARG ARCH
FROM multiarch/alpine:${ARCH}

WORKDIR /opt/clustercode

RUN \
    apk add --no-cache curl ffmpeg && \
    # Let's create the directories first so we can apply the permissions:
    mkdir -m 664 /input /output /clustercode

VOLUME \
    /input \
    /output \
    /clustercode

CMD ["/opt/clustercode/entrypoint.sh"]

COPY ["defaults.yaml", "entrypoint.sh", "./"]
COPY --from=builder /go/src/app/clustercode-worker ./
