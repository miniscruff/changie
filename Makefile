test:
	go test -coverprofile=c.out ./...

test-no-color:
	go test -coverprofile=c.out ./... -ginkgo.noColor -test.failfast

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
	hugo serve -s docs

docs-build: gen
	hugo -s docs --minify --gc

docs-release: gen
	awk 'NR > 1' CHANGELOG.md >> docs/content/guide/changelog/index.md
	hugo --minify -s docs -b https://changie.dev/

preview-changelog:
	go run main.go batch $$(go run main.go next minor)-preview
	go run main.go merge
	awk 'NR > 1' CHANGELOG.md >> docs/content/guide/changelog/index.md
