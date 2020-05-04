#!/bin/bash
# Script is used for deploying new version golang-web-app-demo
#

set -ex

REPO_USERNAME="thaniri"
IMAGE_NAME="golang-web-app-demo"
IMAGE_TAG="latest"

if [[ -z "${DOCKER_USERNAME}" ]]; then
  echo "DOCKER_USERNAME is not set"
  exit 1
fi

if [[ -z "${DOCKER_PASSWORD}" ]]; then
  echo "DOCKER_USERNAME is not set"
  exit 1
fi

echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

# Dockerfile is in the root of the repo
docker build -t "${REPO_USERNAME}"/"${IMAGE_NAME}":"$IMAGE_TAG" .

docker push "${REPO_USERNAME}"/"${IMAGE_NAME}":"$IMAGE_TAG"
