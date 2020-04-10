
.PHONY: test build clean

default: test

build:
	@cd cmd; \
	go build -mod vendor -o nebula-importer; \
	mv nebula-importer ..;
	@echo "nebula-importer has been outputed to $$(pwd)/nebula-importer";

clean:
	rm -rf nebula-importer;

test:
	docker-compose up --exit-code-from importer; \
	docker-compose down -v;
