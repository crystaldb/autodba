#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

docker build -f Dockerfile . --target release --tag autodbarelease
docker run -d --name autodba-release-container autodbarelease tail -f /dev/null
docker cp autodba-release-container:/home/autodba/release_output ./release_output
docker stop autodba-release-container
docker rm autodba-release-container

echo "Release artifacts are available in the release_output directory"
