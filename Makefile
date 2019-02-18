install:
	go install ./blueprint/...
	go install ./go/...

test:
	go test ./go/...

lint:
	golangci-lint run ./gostatic.go
	golangci-lint run ./go/...
	golangci-lint run ./blueprint/main.go
	golangci-lint run ./blueprint/go/...

test-report:
	go test -v -covermode=atomic -coverprofile=coverage.out ./go/...

test-integration:
	bash -c './integration/test.sh'

test-all: test test-integration

mock:
	go install ./vendor/github.com/golang/mock/mockgen
	go generate ./go/...

DOCKER_TAG := gostatic
DOCKER_NAME := $(DOCKER_TAG)
DOCKER_RUN_ARGS := --tty --name $(DOCKER_NAME) $(DOCKER_TAG)
DOCKER_WORKDIR := /go/src/github.com/s12chung/gostatic

docker.build:
	docker build -t $(DOCKER_TAG) --build-arg DOCKER_WORKDIR=$(DOCKER_WORKDIR) .

docker: docker.build
	docker run --rm $(DOCKER_RUN_ARGS) $(WHAT)

docker.dev: docker.build
	docker run --rm -i -v "$$(pwd)":$(DOCKER_WORKDIR):cached $(DOCKER_RUN_ARGS) ash

docker.test-report: docker.build
	docker run $(DOCKER_RUN_ARGS) make test-report
	docker cp $(DOCKER_NAME):$(DOCKER_WORKDIR)/coverage.out ./coverage.out
	make docker.clean

docker.clean:
	docker container rm $(DOCKER_NAME)
