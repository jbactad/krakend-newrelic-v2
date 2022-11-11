deps:
	go get -v -t -d ./...

setup:
	go install github.com/golang/mock/mockgen@v1.6.0
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1

gen:
	go generate

test-unit:
	go test -v .
