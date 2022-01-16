#!/bin/bash
set -e

if [[ "$1" = "serve" ]]; then
    shift 1
    torchserve --start --ts-config /home/model-server/config.properties
else
    eval "$@"
fi

# wait for server to start before registering models
sleep 15
curl --location --request POST 'http://localhost:8081/models?initial_workers=1&synchronous=true&url=cycleganfloraladd.mar'
curl --location --request POST 'http://localhost:8081/models?initial_workers=1&synchronous=true&url=cycleganfloralremove.mar'
curl --location --request POST 'http://localhost:8081/models?initial_workers=1&synchronous=true&url=cycleganstripesadd.mar'
curl --location --request POST 'http://localhost:8081/models?initial_workers=1&synchronous=true&url=cycleganstripesremove.mar'
# prevent docker exit
tail -f /dev/null
