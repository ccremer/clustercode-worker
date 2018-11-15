#!/usr/bin/env bash

repo="${DOCKER_REPOSITORY}"
arch="${DOCKER_ARCH}"
tag="${DOCKER_TAG}"
branch="${TRAVIS_BRANCH}"

# Extract version number from release branch
version="${branch#*origin/}"
echo "=> version: ${version}"
# Push versioned runtime images
echo "=> Logging into Docker Hub..."
echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

new_tag="${version}-${tag}"
docker tag "${repo}:${tag}" "${repo}:${new_tag}"
docker push "${repo}:${new_tag}"

# Push latest images if eligible
latest_branch=$(git branch --remote | grep "\." | sort -r | head -n 1)
if [[ "${branch}" = "${latest_branch#*origin/}" ]]; then
    echo "=> We are on latest release branch, push latest tag"
    docker push "${repo}:${tag}"
    if [[ "${tag}" = "amd64" ]]; then
        docker tag "${repo}:amd64" "${repo}:latest"
        docker push "${repo}:latest"
    fi
fi

# Push master tag if we are on master and on amd64
if [[ "${branch}" = "master" ]] && [[ "${tag}" = "amd64" ]]; then
    echo "=> We are on master branch, push master tag"
    docker tag "${repo}:amd64" "${repo}:master"
    docker push "${repo}:master"
fi


echo "=> Logging out from Docker Hub"
docker logout
