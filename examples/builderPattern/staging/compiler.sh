#!/bin/sh
set -e

echo "Compiling Go Application Image"
echo "------------------------------"

tmp=`mktemp -d`

sed "s/JOB/$job/g" ./staging/Dockerfile.GoCompile > $tmp/Dockerfile.gen

docker build --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy -t $owner/$image_name:$version . -f $tmp/Dockerfile.gen

docker create --name gobuild $owner/$image_name:$version

docker cp gobuild:/go/src/github.com/dgruber/wfl/examples/multi/staging/jobs/$job/job $tmp/job

docker rm -f gobuild

docker rmi $owner/$image_name:$version

mkdir -p ./staging/builds/$job

cp $tmp/job ./staging/builds/$job/job

echo "job binary stored successfully"
