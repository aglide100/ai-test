#!/bin/bash

docker stack deploy -c <(docker-compose -f app.yml config) ai

docker stack deploy -c <(docker-compose -f sdxl.yml config) ai