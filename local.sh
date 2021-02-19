#!/bin/bash
docker rm -f database-cloner
docker build . -t database-cloner
docker run -dit --name database-cloner database-cloner
