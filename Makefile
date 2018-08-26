install:
	dep ensure

test:
	go test ./go/...

mock:
	go install ./vendor/github.com/golang/mock/mockgen
	go generate ./go/...

docker-install: docker-build-install docker-copy

docker-build-install:
	docker-compose up --no-start

# $(shell docker-compose ps -q web) breaks if this target is combined with docker-build-install
DEP_MANAGER_PATHS := vendor Gopkg.lock
docker-copy:
	$(foreach dep_path,$(DEP_MANAGER_PATHS),docker cp $(shell docker-compose ps -q web):$(DOCKER_WORKDIR)/$(dep_path) ./$(dep_path);)

docker-test:
	docker-compose up --exit-code-from web

docker-run-sh:
	docker-compose run web ash