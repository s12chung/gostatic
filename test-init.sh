#!/bin/bash
go install .
${GOPATH}/bin/gostatic init test-init --test
cd test-init
source ./.envrc
make docker-install
docker-compose run web make prod