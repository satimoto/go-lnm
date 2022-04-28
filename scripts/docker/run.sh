#!/bin/sh
docker run -p 9002:9002 -p 50002:50002 --env-file .env.docker lsp