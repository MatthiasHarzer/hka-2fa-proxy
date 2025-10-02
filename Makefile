BUILD_VERSION ?= "unknown"

OUTPUT_NAME := "template"

clean:
	@rm -rf build/

build: clean
	@GOOS=windows GOARCH=amd64 go build -o ./build/$(OUTPUT_NAME).exe -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go
	@GOOS=linux GOARCH=amd64 go build -o ./build/$(OUTPUT_NAME) -ldflags "-X main.version=$(BUILD_VERSION)" ./main.go

qa: analyze test

analyze:
	@go vet
	@go tool staticcheck --checks=all

test:
	@go test -failfast -cover ./...


.PHONY: clean \
				build \
				qa \
				analyze \
				test
