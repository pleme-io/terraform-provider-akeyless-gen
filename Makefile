.PHONY: generate build install test clean

BINARY_NAME=terraform-provider-akeyless-gen
SPEC_PATH ?= $(HOME)/code/github/akeylesslabs/akeyless-go/api/openapi.yaml
RESOURCES_PATH ?= $(HOME)/code/github/pleme-io/akeyless-terraform-resources
OUTPUT_PATH ?= ./internal

generate:
	terraform-forge generate \
		--spec $(SPEC_PATH) \
		--resources $(RESOURCES_PATH)/resources \
		--output $(OUTPUT_PATH) \
		--provider $(RESOURCES_PATH)/provider.toml

build:
	go build -o $(BINARY_NAME)

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/pleme-io/akeyless-gen/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp $(BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/pleme-io/akeyless-gen/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
