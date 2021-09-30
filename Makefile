# Copyright (c) 2020 Cisco Systems Inc. All rights reserved.

BINDIR=$(CURDIR)/.bin
LINTER=$(BINDIR)/golangci-lint
GENERATOR=$(BINDIR)/argen

all: lint test example

check: lint test

$(BINDIR):
	@mkdir -p $@

$(BINDIR)/%: | $(BINDIR)
	env GOBIN=$(BINDIR) go install $(CMD)

$(LINTER): CMD=github.com/golangci/golangci-lint/cmd/golangci-lint

$(GENERATOR): CMD=golang.cisco.com/argo/cmd/argen

# TODO: Fix the hardcoding of argo ddN path.
generate: | $(GENERATOR)
	$(GENERATOR) run -m ../argo/ddN -m ./model -g ./gen 2>$(BINDIR)/argen.err 1>$(BINDIR)/argen.out

lint: | $(LINTER) generate
	$(LINTER) run ./...

clean:
	rm -rf $(BINDIR)
	rm -rf $(CURDIR)/gen
	rm -f example

test: generate
	go test -cover -race ./...

example:
	go build ./cmd/example

.PHONY: generate lint test example
