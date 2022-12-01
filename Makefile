
.PHONY: test build clean

default: build

build: clean fmt
	@cd cmd; \
	CGO_ENABLED=0 go build -ldflags "\
		-X 'github.com/vesoft-inc/nebula-importer/v3/pkg/version.GoVersion=$(shell go version)' \
		-X 'github.com/vesoft-inc/nebula-importer/v3/pkg/version.GitHash=$(shell git rev-parse HEAD)'\
		-X 'github.com/vesoft-inc/nebula-importer/v3/pkg/version.Tag=$(shell git describe --exact-match --abbrev=0 --tags | sed 's/^v//')'\
		" -o nebula-importer; \
	mv nebula-importer ..;
	@echo "nebula-importer has been outputed to $$(pwd)/nebula-importer";

vendor: clean fmt
	@cd cmd; go mod vendor

vendorbuild: vendor
	@cd cmd; \
	CGO_ENABLED=0 go build -mod vendor -o nebula-importer; \
	mv nebula-importer ..;
	@echo "nebula-importer has been outputed to $$(pwd)/nebula-importer";

clean:
	rm -rf nebula-importer;

test:
	docker-compose up --exit-code-from importer; \
	docker-compose down -v;

gotest:
	go test -gcflags=all=-l -race -coverprofile=coverage.txt -covermode=atomic ./...

fmt:
	@go mod tidy && find . -path ./vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;
