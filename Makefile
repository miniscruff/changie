.ONESHELL:
.DEFAULT: help

.PHONY: help
help:
	@grep -E '^[a-z-]+:.*#' Makefile | \
		sort | \
		while read -r l; do printf "\033[1;32m$$(echo $$l | \
		cut -d':' -f1)\033[00m:$$(echo $$l | cut -d'#' -f2-)\n"; \
	done

.PHONY: test
test: # Run unit test suite
	go test -race -coverprofile=c.out ./...
	go tool cover -html=c.out -o=coverage.html

.PHONY: lint
lint: # Run linters
	goimports -w -local github.com/miniscruff/changie .
	golangci-lint run ./...

.PHONY: gen
gen: # Generate config and CLI docs
	go run main.go gen

.PHONY: vhs-gen
vhs-gen: # Generate VHS recording videos
	cd examples
	ls *.tape | xargs -n 1 vhs

.PHONY: docs-serve
docs-serve: gen # Serve documentation locally
	hugo serve -s docs
