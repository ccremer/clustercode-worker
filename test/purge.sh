#!/usr/bin/env bash

docker exec -it rabbitmq rabbitmqctl stop_app
docker exec -it rabbitmq rabbitmqctl reset
docker exec -it rabbitmq rabbitmqctl start_app
