.PHONY: lint
lint: 
	@gometalinter --disable-all --enable=vet --enable=golint --enable=goimports --vendor ./...

.PHONY: test
test:
	@docker run --rm \
		-v $$(pwd):/app \
		-w /app \
		-v $(shell go env GOCACHE):/tmp/gocach \
		-e "GOCACHE=/tmp/gocach" \
		-v $(GOPATH)/pkg/mod:/go/pkg/mod golang:1.12 \
		go test ./...

.PHONY: ci
ci: lint test

.PHONY: dbuild
dbuild:
	@docker build -t terraforming .
