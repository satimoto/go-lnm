#!/bin/sh
docker run -p 9003:9003 -p 50000:50000 --env-file .env.docker lnm