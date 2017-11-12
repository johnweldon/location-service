#!/usr/bin/env bash

IMAGE=docker.jw4.us/location
NAME=location
SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"


docker pull ${IMAGE}
docker stop ${NAME}
docker logs ${NAME} &> $(TZ=UTC date +%Y-%m-%d-%H%M-${NAME}.log)
docker rm -v -f ${NAME}

docker run -d \
  --name ${NAME} \
  --restart=always \
  -e GOOGLE_LOCATION_API_KEY="${GOOGLE_LOCATION_API_KEY}" \
  -e ALL_STORES="${ALL_STORES}" \
  -e STORE_TYPE="${STORE_TYPE}" \
  -p 19981:8182 \
  -v ${SCRIPTDIR}/config:/etc/location \
  -v ${SCRIPTDIR}/certs:/etc/ssl/certs \
  ${IMAGE}

