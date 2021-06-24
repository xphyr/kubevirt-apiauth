.PHONY: image

IMAGE?=quay.io/markd/listvms
GIT_VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)

all: listvms image

image: listvms
	docker build -t $(IMAGE) -f Dockerfile .

listvms: $(shell find . -name "*.go")
	GOOS=linux CGO_ENABLED=0 go build -o listvms .


.PHONY: clean
clean:
	rm listvms
