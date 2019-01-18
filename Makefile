TEST?=$$(go list ./... | grep -v vendor)
VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
GOGEN_FILES?=$$(go list ./... | grep -v vendor)
BIN_NAME?=cloud-plan-migrate
CURRENT_VERSION = $(shell gobump show -r version/)
GO_FILES?=$(shell find . -name '*.go')
export GO111MODULE=on

BUILD_LDFLAGS = "-s -w \
	  -X github.com/sacloud/cloud-plan-migrate/version.Revision=`git rev-parse --short HEAD` \
	  -X github.com/sacloud/cloud-plan-migrate/version.Version=$(CURRENT_VERSION)"

.PHONY: default
default: clean fmt test build

.PHONY: run
run:
	go run $(CURDIR)/main.go $(ARGS)

.PHONY: clean
clean:
	rm -Rf bin/*

.PHONY: clean-all
clean-all:
	rm -Rf bin/*

.PHONY: tools
tools:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/motemen/gobump/cmd/gobump
	go get -u golang.org/x/lint/golint

.PHONY: build build-x build-darwin build-windows build-linux
build: bin/cloud-plan-migrate

bin/cloud-plan-migrate: $(GO_FILES)
	OS="`go env GOOS`" ARCH="`go env GOARCH`" ARCHIVE= BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

build-x: build-darwin build-windows build-linux build-bsd

build-darwin: bin/cloud-plan-migrate_darwin-amd64.zip

build-windows: bin/cloud-plan-migrate_windows-386.zip bin/cloud-plan-migrate_windows-amd64.zip

build-linux: bin/cloud-plan-migrate_linux-386.zip bin/cloud-plan-migrate_linux-amd64.zip bin/cloud-plan-migrate_linux-arm.zip

build-bsd: bin/cloud-plan-migrate_freebsd-386.zip bin/cloud-plan-migrate_freebsd-amd64.zip

bin/cloud-plan-migrate_darwin-amd64.zip:
	OS="darwin"  ARCH="amd64"     ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_windows-386.zip:
	OS="windows" ARCH="386"     ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_windows-amd64.zip:
	OS="windows" ARCH="amd64"     ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_linux-386.zip:
	OS="linux"   ARCH="386" ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_linux-amd64.zip:
	OS="linux"   ARCH="amd64" ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_linux-arm.zip:
	OS="linux"   ARCH="arm" ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_freebsd-386.zip:
	OS="freebsd"   ARCH="386" ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

bin/cloud-plan-migrate_freebsd-amd64.zip:
	OS="freebsd"   ARCH="amd64" ARCHIVE=1 BUILD_LDFLAGS=$(BUILD_LDFLAGS) sh -c "'$(CURDIR)/scripts/build.sh'"

.PHONY: test
test: 
	go test $(TEST) $(TESTARGS) -v -timeout=30m -parallel=4 ;

.PHONY: integration-test
integration-test: bin/cloud-plan-migrate
	test/integration/run-bats.sh test/integration/bats ;

.PHONY: lint
lint: golint
	gometalinter --vendor --skip=vendor/ --disable-all --enable vet --enable goimports --deadline=5m ./...
	@echo

.PHONY: golint
golint: goimports
	for pkg in $$(go list ./... | grep -v /vendor/ ) ; do \
        test -z "$$(golint $$pkg | grep -v '_gen.go' | grep -v '_string.go' | grep -v 'should have comment' | grep -v 'func ServerMonitorCpu' | grep -v 'func ServerSsh' | grep -v 'DatabaseMonitorCpu' | grep -v "func MobileGatewayDnsUpdate" | tee /dev/stderr)" || RES=1; \
    done ;exit $$RES

.PHONY: goimports
goimports:
	goimports -l -w $(GOFMT_FILES)

.PHONY: fmt
fmt:
	gofmt -s -l -w $(GOFMT_FILES)

.PHONY: build-docs serve-docs lint-docs
build-docs:
	sh -c "'$(CURDIR)/scripts/build_docs.sh'"

serve-docs:
	sh -c "'$(CURDIR)/scripts/serve_docs.sh'"

lint-docs:
	sh -c "'$(CURDIR)/scripts/lint_docs.sh'"

.PHONY: docker-run docker-test docker-build
docker-run:
	sh -c "$(CURDIR)/scripts/build_docker_image.sh" ; \
	sh -c "$(CURDIR)/scripts/run_on_docker.sh"

docker-test:
	sh -c "'$(CURDIR)/scripts/build_on_docker.sh' 'test'"

docker-integration-test:
	sh -c "'$(CURDIR)/scripts/run_integration_test.sh'"

docker-build: clean
	sh -c "'$(CURDIR)/scripts/build_on_docker.sh' 'build-x'"

.PHONY: bump-patch bump-minor bump-major version
bump-patch:
	@gobump patch -w version/

bump-minor:
	@gobump minor -w version/

bump-major:
	@gobump major -w version/

version:
	@gobump show -r version/

git-tag:
	git tag v`gobump show -r version`
