
.PHONY: test build clean

default: build

build: clean fmt
	@cd cmd; \
	CGO_ENABLED=0 go build -o nebula-importer; \
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

fmt:
	@go mod tidy && find . -path ./vendor -prune -o -type f -iname '*.go' -exec go fmt {} \;
