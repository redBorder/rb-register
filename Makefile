build:
	go build -o rb-register ./cmd/app/

check: fmt errcheck vet

fmt:
	@if [ -n "$$(go fmt ./...)" ]; then echo 'Please run go fmt on your code.' && exit 1; fi

errcheck:
	errcheck -ignoretests -verbose ./...

vet:
	go vet ./...

test:
	go test -cover ./...

get_dev_deps:
	go get golang.org/x/tools/cmd/cover
	go get golang.org/x/tools/cmd/vet
	go get github.com/kisielk/errcheck
	go get github.com/stretchr/testify/assert

get_deps:
	go get -t ./...