#!/usr/bin/env bash

repo="${DOCKER_REPOSITORY}"
tag="${DOCKER_TAG}"
arch="${DOCKER_ARCH}"

yell() { echo "$0: $*" >&2; }
die() { yell "$*"; exit 111; }
try() { "$@" || die "cannot $*"; }

# Install cross-build libraries
try docker run --rm --privileged multiarch/qemu-user-static:register --reset

# Build builder image
try docker build --tag "${repo}:builder" --file ./builder.Dockerfile ./

# Build runtime images
try docker build --build-arg ARCH="${arch}" --tag "${repo}:${tag}" --file ./Dockerfile ./
