test:
	go test -coverprofile=c.out ./...

coverage: test
	go tool cover -html=c.out

watch:
	ginkgo watch ./... -failFast

lint:
	golangci-lint run ./...

format:
	gofmt -s -w .
	goimports -w -local github.com/miniscruff/changie .

gen-cli-docs:
	go run main.go gen

docs-serve:
	hugo serve -s website

docs-build: gen-cli-docs
	hugo -s website --minify --gc
