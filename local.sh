#!/bin/bash
docker rm -f database-cloner
make
docker build . -t database-cloner
docker run -dit -v ${PWD}/config.yaml:/config.yaml --name database-cloner database-cloner
