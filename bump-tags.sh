#!/bin/bash

TAG=$1
DRIVERS=("mbim" "qmi" "ccid")

if [ -z $TAG ]; then
    echo "Usage: bump-tags.sh <tag>"
    exit 1
fi

for driver in ${DRIVERS[@]}; do
    echo "Bumping $driver to $TAG"
    git tag -as "driver/$driver/$TAG" -m "Bump $driver to $TAG"
    git push origin "driver/$driver/$TAG"
done

