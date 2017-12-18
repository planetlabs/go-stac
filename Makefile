.PHONY: all bin clean

all: bin

bin: target/stac_linux_amd64

clean:
	rm -fr target

# sources are anything that can change that will change the build output for
# target/stac_linux_amd64. This is simply anything not in the target directory
# and not in the hidden .
SRCS = $(shell find . -type f -not -path './target/*' -not -path './.*')

# build stac with golang 1.9.2 for linux_amd64
target/stac_linux_amd64: $(SRCS)
	docker run \
		-e "GOPATH=/go" \
		-e "GOOS=linux" \
		-e "GOARCH=amd64" \
		--rm \
		-v "$(shell pwd)":/go/src/github.com/planetlabs/go-stac \
		-w /go/src/github.com/planetlabs/go-stac \
		golang:1.9.2-stretch \
		go build -v -o $@ main.go
