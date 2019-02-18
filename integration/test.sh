#!/bin/bash
go install .
gostatic new test-integration --test
cd test-integration
source ./.envrc
make docker-install
docker-compose run web make prod
