#!/bin/bash

DRIVER=$1
TAG=$2
DRIVERS=("mbim" "qmi" "ccid" "at")

if [ "$#" -ne 2 ]; then
    echo "Usage: bump-tags.sh <driver> <tag>"
    exit 1
fi

if [[ ! " ${DRIVERS[@]} " =~ " ${DRIVER} " ]]; then
    echo "Invalid driver: $DRIVER"
    exit 1
fi

echo "Bumping $DRIVER to $TAG"
git tag -as "driver/$DRIVER/$TAG" -m "Bump $DRIVER to $TAG"
git push origin "driver/$DRIVER/$TAG"
