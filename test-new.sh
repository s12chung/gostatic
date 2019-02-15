#!/bin/bash
go install .
gostatic new test-new --test
cd test-new
source ./.envrc
make docker-install
docker-compose run web make prod
