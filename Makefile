ifndef GOPATH
	GOPATH := $(HOME)/go
endif

ifndef GOOS
	GOOS := linux
endif

ifndef GO111MODULE
	GO111MODULE := on
endif

PREFIX ?= "/opt/cluster"

VERSION ?= "master"

CACERTDIR ?= "/etc/pki/tls/certs"

GOLDFLAGS = "-X github.com/Donders-Institute/hpc-utility/internal/cmd.defTorqueHelperCert=/etc/pki/tls/certs/star_dccn_nl.chained.crt \
-X github.com/Donders-Institute/hpc-utility/internal/cmd.defWebhookCert=/etc/pki/tls/certs/star_dccn_nl.chained.crt \
-X github.com/Donders-Institute/hpc-utility/internal/cmd.defMachineListFile=/opt/cluster/etc/machines.mentat \
-X github.com/Donders-Institute/hpc-utility/internal/cmd.defVersion=$(VERSION)"

.PHONY: build

all: build

build:
	GOPATH=$(GOPATH) GOOS=$(GOOS) GO111MODULE=$(GO111MODULE) GOLDFLAGS=$(GOLDFLAGS) go install -ldflags $(GOLDFLAGS) github.com/Donders-Institute/hpc-utility/...

doc:
	@GOPATH=$(GOPATH) GOOS=$(GOOS) godoc -http=:6060

test: build_dep
	@GOPATH=$(GOPATH) GOOS=$(GOOS) GOCACHE=off go test \
	-ldflags $(GOLDFLAGS) -v github.com/Donders-Institute/hpc-utility/...

install: build
	@install -D $(GOPATH)/bin/* $(PREFIX)/bin

release:
	VERSION=$(VERSION) rpmbuild --undefine=_disable_source_fetch -bb build/rpm/centos7.spec

github_release:
	scripts/gh-release.sh $(VERSION) false

clean:
	@rm -rf $(GOPATH)/bin/cluster-*
	@rm -rf $(GOPATH)/pkg/*/Donders-Institute/hpc-utility
	@rm -rf $(GOPATH)/pkg/*/Donders-Institute/hpc-utility
