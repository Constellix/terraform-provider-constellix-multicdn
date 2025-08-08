HOSTNAME=registry.terraform.io
NAMESPACE=constellix
NAME=constellix-multicdn
VERSION := $(shell cat VERSION)
OS_ARCH=darwin_arm64

default: install

.PHONY: build
build:
	go build -o terraform-provider-${NAME}_v${VERSION}

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp terraform-provider-${NAME}_v${VERSION} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/terraform-provider-${NAME}_v${VERSION}

.PHONY: test
test:
	go test -v ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v ./provider/... -timeout 120m

.PHONY: clean
clean:
	rm -f terraform-provider-${NAME}_v${VERSION}

.PHONY: docs
docs:
	./gendoc.sh
