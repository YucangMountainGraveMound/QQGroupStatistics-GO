.PHONY: test

PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

install:
	glide install

update:
	glide up

test:
	for PKG in $(PACKAGES); do go test -v -cover -coverprofile $$GOPATH/src/$$PKG/coverage.txt $$PKG || exit 1; done;

html:
	go tool cover -html=.cover/coverage.txt

fmt:
	find . -name "*.go" -type f -not -path "./vendor/*" | xargs gofmt -s -w

vet:
	go vet $(PACKAGES)

errcheck:
	@which errcheck > /dev/null; if [ $$? -ne 0 ]; then \
		go get -u github.com/kisielk/errcheck; \
	fi
	errcheck $(PACKAGES)

lint:
	@which golint > /dev/null; if [ $$? -ne 0 ]; then \
		go get -u github.com/golang/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

coverage:
	curl -s https://codecov.io/bash > .codecov && \
	chmod +x .codecov && \
	./.codecov -f .cover/coverage.txt

clean:
	rm -rf .cover
	find . -name "coverage.txt"
