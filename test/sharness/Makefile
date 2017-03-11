# Run sharness tests
#
# Copyright (c) 2014 Christian Couder
# MIT Licensed; see the LICENSE file in this repository.
#
# NOTE: run with TEST_VERBOSE=1 for verbose sharness tests.

T = $(sort $(wildcard t[0-9][0-9][0-9][0-9]-*.sh))
LIBDIR = lib
SHARNESSDIR = sharness
AGGREGATE = $(LIBDIR)/$(SHARNESSDIR)/aggregate-results.sh

find_go_files = $(shell find $(1) -name "*.go")

BINS = bin/ipfs-pack

# User might want to override those on the command line
GOFLAGS =

all: aggregate

clean: clean-test-results
	@echo "*** $@ ***"
	-rm -rf $(BINS)

clean-test-results:
	@echo "*** $@ ***"
	-rm -rf test-results

$(T): clean-test-results deps
	@echo "*** $@ ***"
	./$@

aggregate: clean-test-results $(T)
	@echo "*** $@ ***"
	ls test-results/t*-*.sh.*.counts | $(AGGREGATE)

deps: sharness $(BINS)

sharness:
	@echo "*** checking $@ ***"
	lib/install-sharness.sh

bin/ipfs-pack: $(call find_go_files, ../../)
	mkdir -p bin
	go build -o bin/ipfs-pack $(GOFLAGS) github.com/ipfs/ipfs-pack

.PHONY: all clean clean-test-results $(T) aggregate deps sharness FORCE

