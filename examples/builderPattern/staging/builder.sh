#!/bin/sh
set -e

echo "Building Binary Application Image"
echo "---------------------------------"

tmp=`mktemp -d`

sed "s/APP/$job/g" ./staging/Dockerfile.GoBuild > $tmp/Dockerfile.gen

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy -t $owner/$image_name:$version . -f $tmp/Dockerfile.gen
