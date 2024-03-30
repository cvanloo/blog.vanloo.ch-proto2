.PHONY: all serve

default: all

SOURCES := $(shell find . -name '*.go')

be: $(SOURCES)
	go build cmd/be.go

all: be
	./be

serve: be
	./be -serve=true
