test:
	go test -coverprofile=c.out ./...

coverage: test
	go tool cover -html=c.out

watch:
	ginkgo watch ./... -failFast

lint:
	gofmt -s -w .
	goimports -w -local github.com/miniscruff/changie .
	golangci-lint run ./...

gen-cli-docs:
	go run main.go gen

docs-serve:
	hugo serve -s docs

docs-build: gen-cli-docs
	hugo -s docs --minify --gc

docs-release: gen-cli-docs
	awk 'NR > 1' CHANGELOG.md >> docs/content/guide/changelog/index.md
	hugo --minify -s docs -b https://changie.dev/

preview-changelog:
	go run main.go batch $$(go run main.go next minor)-preview
	go run main.go merge
	awk 'NR > 1' CHANGELOG.md >> docs/content/guide/changelog/index.md
