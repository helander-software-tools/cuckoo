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
docker inspect $CONTAINER|jq '.[].Config.WorkingDir'  > .cuckoo/dir
docker inspect $CONTAINER|jq '.[].Config.Env'         > .cuckoo/env
docker inspect $CONTAINER|jq '.[].Config.Entrypoint'  > .cuckoo/entrypoint
docker inspect $CONTAINER|jq '.[].Config.Cmd'         > .cuckoo/cmd

docker cp .cuckoo $CONTAINER:/

docker export $CONTAINER | gzip > ${COMPONENT}_${PLATFORM}.tar.gz 

docker rm -f $CONTAINER

rm -fr .cuckoo
