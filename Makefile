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

gen:
	go run main.go gen

docs-serve:
	hugo serve -s website

docs-build: gen
	hugo -s website --minify --gc

docs-release: gen
	awk 'NR > 1' CHANGELOG.md >> website/content/guide/changelog/index.md
	hugo --minify -s website -b https://changie.dev/

preview-changelog:
	go run main.go batch $$(go run main.go next minor)-preview
	go run main.go merge
	awk 'NR > 1' CHANGELOG.md >> website/content/guide/changelog/index.md
