########################################################################################################################
#
# VARS THAT CAN BE EDITED
#
########################################################################################################################
# Project metadata
GOVERSION := 1.11
PROJECT := github.com/ekino/terraform-provider-greseller
NAME := $(notdir $(PROJECT))
VERSION ?= 0.0.0

# Path to Terraform plugins for devevelopment tests
PLUGIN_PATH ?= "${HOME}/.terraform.d/plugins"

# Default os-arch combination to build
XC_OS ?= linux darwin freebsd openbsd solaris windows
XC_ARCH ?= amd64 386 arm
XC_EXCLUDE ?= darwin/386 darwin/arm solaris/386 solaris/arm windows/arm

########################################################################################################################
#
# VARS THAT SHOULD NOT BE EDITED
#
########################################################################################################################
# Metadata about this makefile and position
MKFILE_PATH := $(lastword $(MAKEFILE_LIST))
CURRENT_DIR := $(patsubst %/,%,$(dir $(realpath $(MKFILE_PATH))))

# Build flags
LD_FLAGS ?= -s -w
GOTAGS ?=

# List of tests to run
TEST ?= ./...
TESTARGS ?=

# GPG Signing key (blank by default, means no GPG signing)
GPG_KEY ?=

# COLORS
RED    = $(shell printf "\33[31m")
GREEN  = $(shell printf "\33[32m")
WHITE  = $(shell printf "\33[37m")
YELLOW = $(shell printf "\33[33m")
RESET  = $(shell printf "\33[0m")


########################################################################################################################
#
# DEVELOPMENT PROCESS
#
########################################################################################################################

dev: ##@dev build plugin and install in terraform plugin path for tests
	@mkdir -p "${PLUGIN_PATH}"
	@go build \
		-ldflags "${LD_FLAGS}" \
		-tags "${GOTAGS}" \
		-o "${PLUGIN_PATH}/terraform-provider-greseller"
.PHONY: dev

deps: ##@dev Install dependencies
	@dep ensure -v
	@dep prune
.PHONY: deps

deps-update: ##@dev Update dependencies
	@dep ensure -v -update
	@dep prune
.PHONY: deps-update

test: ##@dev runs all tests
	@echo "${RED}There is currently no tests. TODO${RESET}"
	#go test -i $(TEST) || exit 1
	#echo $(TEST) | \
	#	xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4
.PHONY: test

########################################################################################################################
#
# BUILD PROCESS
#
########################################################################################################################

# Create a cross-compile target for every os-arch pairing. This will generate
# a make target for each os/arch like "make linux/amd64" as well as generate a
# meta target (build) for compiling everything.
define make-xc-target
  $1/$2:
  ifneq (,$(findstring ${1}/${2},$(XC_EXCLUDE)))
		@echo "${RED}Build for platform ${1}/${2} was excluded${RESET}"
  else
		@echo "${YELLOW}Building for platform ${1}/${2}${RESET}"
		@docker run \
			--rm \
			--volume="${CURRENT_DIR}:/go/src/${PROJECT}" \
			--workdir="/go/src/${PROJECT}" \
			"golang:${GOVERSION}" \
			env \
				CGO_ENABLED="0" \
				GOOS="${1}" \
				GOARCH="${2}" \
				go build \
				  -a \
					-o="pkg/${1}_${2}/${NAME}_v${VERSION}${3}" \
					-ldflags "${LD_FLAGS}" \
					-tags "${GOTAGS}"
  endif
  .PHONY: $1/$2

  $1:: $1/$2
  .PHONY: $1

  build:: $1/$2
  .PHONY: build
endef
$(foreach goarch,$(XC_ARCH),$(foreach goos,$(XC_OS),$(eval $(call make-xc-target,$(goos),$(goarch),$(if $(findstring windows,$(goos)),.exe,)))))

########################################################################################################################
#
# RELEASE PROCESS
#
########################################################################################################################
dist: ##@dist build for all platforms, compress and compute checksum
	@$(MAKE) -f "${MKFILE_PATH}" clean
	@$(MAKE) -f "${MKFILE_PATH}" -j4 build
	@$(MAKE) -f "${MKFILE_PATH}" compress checksum
.PHONY: dist

clean: ##@release removes any previous binaries
	@echo "${YELLOW}Clean build artifacts${RESET}"
	@rm -rf "${CURRENT_DIR}/pkg/"
	@rm -rf "${CURRENT_DIR}/bin/"
.PHONY: clean

compress: ##@release compresses all the binaries in pkg/* as tarball and zip
	@echo "${YELLOW}Compress binaries for release${RESET}"
	@mkdir -p "${CURRENT_DIR}/pkg/dist"
	@for platform in $$(find ./pkg -mindepth 1 -maxdepth 1 -type d); do \
		osarch=$$(basename "$$platform"); \
		if [ "$$osarch" = "dist" ]; then \
			continue; \
		fi; \
		\
		ext=""; \
		if test -z "$${osarch##*windows*}"; then \
			ext=".exe"; \
		fi; \
		cd "$$platform"; \
		tar -czf "${CURRENT_DIR}/pkg/dist/${NAME}_${VERSION}_$${osarch}.tgz" "${NAME}_v${VERSION}$${ext}"; \
		zip -q "${CURRENT_DIR}/pkg/dist/${NAME}_${VERSION}_$${osarch}.zip" "${NAME}_v${VERSION}$${ext}"; \
		cd - &>/dev/null; \
	done
.PHONY: compress

checksum: ##@release produces the checksums for compressed binaries
	@echo "${YELLOW}Compute checksum${RESET}"
	@cd "${CURRENT_DIR}/pkg/dist" && \
		shasum --algorithm 256 * > ${CURRENT_DIR}/pkg/dist/${NAME}_${VERSION}_SHA256SUMS && \
		cd - &>/dev/null
.PHONY: checksum

sign: ##@release sign checksum using the given GPG_KEY.
	@echo "${YELLOW}Signing checksum${RESET}"
ifndef GPG_KEY
	@echo "${RED}ERROR: No GPG key specified! Without a GPG key, this release cannot"
	@echo "           be signed. Set the environment variable GPG_KEY to the ID of"
	@echo "           the GPG key to continue.${RESET}"
	@exit 127
else
	@gpg \
		--default-key "${GPG_KEY}" \
		--detach-sig "${CURRENT_DIR}/pkg/dist/${NAME}_${VERSION}_SHA256SUMS"
endif
.PHONY: sign

tag: ##@release sign commit and tag version
	@echo "${YELLOW}Signing commit${RESET}"
ifndef GPG_KEY
	@echo "${RED}ERROR: No GPG key specified! Without a GPG key, this release cannot"
	@echo "           be signed. Set the environment variable GPG_KEY to the ID of"
	@echo "           the GPG key to continue.${RESET}"
	@exit 127
else
	@git commit \
		--allow-empty \
		--gpg-sign="${GPG_KEY}" \
		--message "Release v${VERSION}" \
		--quiet \
		--signoff
	@git tag \
		--annotate \
		--create-reflog \
		--local-user "${GPG_KEY}" \
		--message "Version ${VERSION}" \
		--sign \
		"v${VERSION}"
	@echo "${YELLOW}Do not forget to run:"
	@echo ""
	@echo "    git push && git push --tags"
	@echo ""
	@echo "And then upload the binaries in dist/!${RESET}"
endif
.PHONY: tag

########################################################################################################################
#
# HELP
#
########################################################################################################################

# Add the following 'help' target to your Makefile
# And add help text after each target name starting with '\#\#'
# A category can be added with @category
HELP_HELPER = \
    %help; \
    while(<>) { push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z\-\%]+)\s*:.*\#\#(?:@([a-zA-Z\-\%]+))?\s(.*)$$/ }; \
    print "usage: make [target]\n\n"; \
    for (sort keys %help) { \
    print "${WHITE}$$_:${RESET}\n"; \
    for (@{$$help{$$_}}) { \
    $$sep = " " x (32 - length $$_->[0]); \
    print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
    }; \
    print "\n"; }

help: ##prints help
	@perl -e '$(HELP_HELPER)' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help