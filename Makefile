BIN_DIR := $(GOPATH)/bin

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLINT := $(BIN_DIR)/golint

.PHONY: help
help: Makefile ## This help dialog
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "%-30s %s\n" $$help_command $$help_info ; \
	done

$(GOLANGCI_LINT):
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOLINT):
	@go get -u golang.org/x/lint/golint

.PHONY: lint
lint: $(GOLANGCI_LINT) $(GOLINT) ## Runs the linter
	@GO111MODULE=on golangci-lint run -E goimports ./... && golint -set_exit_status ./...

.PHONY: test
test: ## Runs the tests
	@docker run --rm \
		-v $$(pwd):/app \
		-w /app \
		-v $(shell go env GOCACHE):/tmp/gocach \
		-e "GOCACHE=/tmp/gocach" \
		-v $(GOPATH)/pkg/mod:/go/pkg/mod golang:1.12 \
		go test ./...

.PHONY: ci
ci: lint test ## Runs the linter and the tests

.PHONY: dbuild
dbuild: ## Builds the docker image with name 'terraforming'
	@docker build -t terraforming .
