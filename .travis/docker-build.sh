#!/usr/bin/env bash

repo="${DOCKER_REPOSITORY}"
tag="${DOCKER_TAG}"
arch="${DOCKER_ARCH}"

# Install cross-build libraries
docker run --rm --privileged multiarch/qemu-user-static:register --reset

# Build builder image
docker build --tag "${repo}:builder" --file ./builder.Dockerfile ./

# Build runtime images
docker build --build-arg ARCH="${arch}" --tag "${repo}:${tag}" --file ./Dockerfile ./
