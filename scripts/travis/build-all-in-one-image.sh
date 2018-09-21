#!/bin/bash

set -e

BRANCH=${BRANCH:?'missing BRANCH env var'}

source ~/.nvm/nvm.sh
nvm use 6
make build-all-in-one-linux

export REPO=jaegertracing/all-in-one

docker build -f cmd/all-in-one/Dockerfile -t $REPO:latest .
export CID=$(docker run -d -p 16686:16686 -p 5778:5778 $REPO:latest)
make integration-test
docker kill $CID

