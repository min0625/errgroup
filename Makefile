MODULE_DIRS = .

gowork:
	go work init .

tidy:
	go mod tidy

install-asdf:
	-asdf install

install: install-asdf tidy
# 	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.5.0

fmt: install
	golangci-lint fmt -v ./...

fix: install
	golangci-lint run -v --fix ./...

lint: install
	golangci-lint run -v ./...

test: install
	go test -v -race -failfast ./...

check: fix lint test
