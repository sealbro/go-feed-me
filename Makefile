#-include .env

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
all, help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nMakefile help:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

vendor:  ### vendor
	go mod tidy
	go mod vendor
.PHONY: vendor

generate: generate_gql ### generate all
	echo "generate..."
.PHONY: generate

generate_gql: ### generate graphql
	cd ./graph
	go run github.com/99designs/gqlgen generate
.PHONY: generate_gql

deps: ### install deps
	brew install bufbuild/buf/buf protobuf
.PHONY: deps
