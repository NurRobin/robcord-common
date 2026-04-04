SHELL := /bin/bash

.PHONY: gen-proto proto-lint

gen-proto: ## Generate Go code from proto schema
	@echo "── Generating proto code ──"
	@cd proto && buf generate
	@echo "── Proto generation complete ──"

proto-lint: ## Lint proto files
	@cd proto && buf lint
