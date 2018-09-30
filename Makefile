install:
	dep ensure

test:
	go test ./go/...

lint:
	golangci-lint run ./gostatic.go
	golangci-lint run ./go/...
	golangci-lint run ./blueprint/main.go
	golangci-lint run ./blueprint/go/...

test-report:
	go test -v -covermode=atomic -coverprofile=coverage.out ./go/...

test-init:
	bash -c './test-init.sh'

test-all: test test-init

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

docker-test-report: docker-test-report-run docker-test-report-copy

docker-test-report-run:
	docker-compose -f docker-compose.yml -f docker-compose.report.yml up --exit-code-from web

# $(shell docker-compose ps -q web) breaks if this target is combined with docker-rest-report-run
docker-test-report-copy:
	docker cp $(shell docker-compose ps -q web):$(DOCKER_WORKDIR)/coverage.out ./coverage.out

docker-run-sh:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml run web ash
