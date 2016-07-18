#!/bin/bash -x

BASE_IMAGE_PATH=$1
TMP_DIR=$2

MOUNT_POINT=$(mktemp -d)

mount -t aufs -o br="$TMP_DIR=rw:$BASE_IMAGE_PATH=r" none $MOUNT_POINT

echo $MOUNT_POINT
