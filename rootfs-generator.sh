#!/bin/bash

COMPONENT_REGISTRY=$2
COMPONENT_PLATFORM=$1

if [ "$COMPONENT_REGISTRY" == "" ];then 
   echo "Missing arguments, to run use: rootfs-generator.sh <platform> <registry>  , e.g.  rootfs-generator linux/arm64 nodered/node-red"
   exit 1 
fi

COMPONENT=$(echo $COMPONENT_REGISTRY|tr '/' '_'|tr ':' '_')
PLATFORM=$(echo $COMPONENT_PLATFORM|tr '/' '_')

CONTAINER=$(docker create --platform $COMPONENT_PLATFORM $COMPONENT_REGISTRY) 

rm -fr .cuckoo
mkdir .cuckoo
docker inspect --format '{{.Config.WorkingDir}}' $CONTAINER > .cuckoo/dir 
docker inspect --format '{{.Config.Env}}' $CONTAINER |tr '[' ' '|tr ']' ' '|xargs|tr ' ' '\n' > .cuckoo/env
docker inspect --format '{{.Config.Entrypoint}}' $CONTAINER |tr '[' ' '|tr ']' ' '|xargs|tr ' ' '\n' > .cuckoo/args
if [ "$(cat .cuckoo/args)" == "" ];then
   docker inspect --format '{{.Config.Cmd}}' $CONTAINER |tr '[' ' '|tr ']' ' '|xargs|tr ' ' '\n' > .cuckoo/args
fi

docker cp .cuckoo $CONTAINER:/

docker export $CONTAINER | gzip > ${COMPONENT}_${PLATFORM}.tar.gz 

docker rm -f $CONTAINER

rm -fr .cuckoo
