.DEFAULT_GOAL := help

SHELL := /bin/bash


gomodule:  ## Tidy up Golang dependencies
	go mod tidy


test:  ## Test to all of directories
	go test -v -cover -race ./...


help:  ## Show all of tasks
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


