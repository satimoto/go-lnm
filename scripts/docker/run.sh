#!/bin/sh
docker run -p 9002:9002 -p 50000:50000 --env-file .env.docker lsp