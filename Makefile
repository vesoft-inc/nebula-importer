DOCKER_REGISTRY ?= localhost:5000
DOCKER_REPO ?= ${DOCKER_REGISTRY}/vesoft
IMAGE_TAG ?= latest

export GO111MODULE := on
GOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=0
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

go-generate: $(GOBIN)/mockgen
	go generate ./...

check: tidy fmt vet imports lint

tidy:
	go mod tidy

fmt: $(GOBIN)/gofumpt
	# go fmt ./...
	$(GOBIN)/gofumpt -w -l ./

vet:
	go vet ./...

imports: $(GOBIN)/goimports $(GOBIN)/impi
	$(GOBIN)/impi --local github.com/vesoft-inc --scheme stdLocalThirdParty \
	    -ignore-generated ./... \
	    || exit 1

lint: $(GOBIN)/golangci-lint
	$(GOBIN)/golangci-lint run

build:
	$(GO_BUILD) -ldflags '$(LDFLAGS)' -o bin/nebula-importer ./cmd/nebula-importer/

test:
	go test -gcflags=all="-l" -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...

docker-build:
	docker build -t "${DOCKER_REPO}/nebula-importer:${IMAGE_TAG}" -f Dockerfile .

docker-push: docker-build
	docker push "${DOCKER_REPO}/nebula-importer:${IMAGE_TAG}"

tools: $(GOBIN)/goimports \
	$(GOBIN)/impi \
	$(GOBIN)/gofumpt \
	$(GOBIN)/golangci-lint \
	$(GOBIN)/mockgen

$(GOBIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@v0.1.12

$(GOBIN)/impi:
	go install github.com/pavius/impi/cmd/impi@v0.0.3

$(GOBIN)/gofumpt:
	go install mvdan.cc/gofumpt@v0.3.1

$(GOBIN)/golangci-lint:
	@[ -f $(GOBIN)/golangci-lint ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.49.0 ;\
	}

$(GOBIN)/mockgen:
	go install github.com/golang/mock/mockgen@v1.6.0
