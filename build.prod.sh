#!/usr/bin/env bash
set -e

VERSION=0.0.1
IMAGE=ghcr.io/teiddy0207/pipevid-transcode-worker:$VERSION

docker build -f Dockerfile.worker -t $IMAGE . --progress=plain

docker push $IMAGE

docker image rm $IMAGE
