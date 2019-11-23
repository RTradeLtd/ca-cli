all: build lint test

.PHONY: all

VERSION="v1.0.0-rt"

-include make/common.mk
-include make/docker.mk

#########################################
# Debian
#########################################

changelog:
	$Q echo "step-cli ($(VERSION)) unstable; urgency=medium" > debian/changelog
	$Q echo >> debian/changelog
	$Q echo "  * See https://github.com/RTradeLtd/ca-cli/releases" >> debian/changelog
	$Q echo >> debian/changelog
	$Q echo " -- Smallstep Labs, Inc. <techadmin@smallstep.com>  $(shell date -uR)" >> debian/changelog

debian: changelog
	$Q set -e; mkdir -p $(RELEASE); \
	OUTPUT=../step-cli_*.deb; \
	rm -f $$OUTPUT; \
	dpkg-buildpackage -b -rfakeroot -us -uc && cp $$OUTPUT $(RELEASE)/

distclean: clean

.PHONY: changelog debian distclean

#################################################
# Build statically compiled step binary for various operating systems
#################################################

BINARY_OUTPUT=$(OUTPUT_ROOT)binary/
BUNDLE_MAKE=v=$v GOOS_OVERRIDE='GOOS=$(1) GOARCH=$(2)' PREFIX=$(3) make $(3)bin/step
RELEASE=./.travis-releases

binary-linux:
	$(call BUNDLE_MAKE,linux,amd64,$(BINARY_OUTPUT)linux/)

binary-darwin:
	$(call BUNDLE_MAKE,darwin,amd64,$(BINARY_OUTPUT)darwin/)

define BUNDLE
	$(q)set -e; BUNDLE_DIR=$(BINARY_OUTPUT)$(1)/bundle; \
	stepName=step_$(2); \
 	mkdir -p $$BUNDLE_DIR $(RELEASE); \
	TMP=$$(mktemp -d $$BUNDLE_DIR/tmp.XXXX); \
	trap "rm -rf $$TMP" EXIT INT QUIT TERM; \
	newdir=$$TMP/$$stepName; \
	mkdir -p $$newdir/bin; \
	cp $(BINARY_OUTPUT)$(1)/bin/step $$newdir/bin/; \
	cp README.md $$newdir/; \
	NEW_BUNDLE=$(RELEASE)/step_$(2)_$(1)_$(3).tar.gz; \
	rm -f $$NEW_BUNDLE; \
    tar -zcvf $$NEW_BUNDLE -C $$TMP $$stepName;
endef

bundle-linux: binary-linux
	$(call BUNDLE,linux,$(VERSION),amd64)

bundle-darwin: binary-darwin
	$(call BUNDLE,darwin,$(VERSION),amd64)

.PHONY: binary-linux binary-darwin bundle-linux bundle-darwin

#################################################
# Targets for creating OS specific artifacts and archives
#################################################

artifacts-linux-tag: bundle-linux debian

artifacts-darwin-tag: bundle-darwin

artifacts-archive-tag:
	$Q mkdir -p $(RELEASE)
	$Q git archive v$(VERSION) | gzip > $(RELEASE)/step-cli_$(VERSION).tar.gz

artifacts-tag: artifacts-linux-tag artifacts-darwin-tag artifacts-archive-tag

.PHONY: artifacts-linux-tag artifacts-darwin-tag artifacts-archive-tag artifacts-tag

#################################################
# Targets for creating step artifacts
#################################################

# For all builds that are not tagged
artifacts-master:

# For all builds with a release candidate tag
artifacts-release-candidate: artifacts-tag

# For all builds with a release tag
artifacts-release: artifacts-tag

# This command is called by travis directly *after* a successful build
artifacts: artifacts-$(PUSHTYPE) docker-$(PUSHTYPE)

.PHONY: artifacts-master artifacts-release-candidate artifacts-release artifacts


# run standard go tooling for better rcode hygiene
.PHONY: tidy
tidy: imports fmtt
	go vet ./...
	golint ./...

# automatically add missing imports
.PHONY: imports
imports:
	find . -type f -name '*.go' -exec goimports -w {} \;

# format code and simplify if possible
.PHONY: fmtt
fmtt:
	find . -type f -name '*.go' -exec gofmt -s -w {} \;