PREFIX ?= "/opt/cluster"

VERSION ?= "master"

GOOS ?= "linux"

CACERTDIR ?= "/etc/pki/tls/certs"

GOLDFLAGS = "-X github.com/Donders-Institute/hpc-utility/internal/cmd.defTorqueHelperCert=/etc/pki/tls/certs/star_dccn_nl.chained.crt \
-X github.com/Donders-Institute/hpc-utility/internal/cmd.defWebhookCert=/etc/pki/tls/certs/star_dccn_nl.chained.crt \
-X github.com/Donders-Institute/hpc-utility/internal/cmd.defVersion=$(VERSION)"

all: build

$(GOPATH)/bin/dep:
	mkdir -p $(GOPATH)/bin
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | GOPATH=$(GOPATH) GOOS=linux sh

build_dep: $(GOPATH)/bin/dep
	GOPATH=$(GOPATH) GOOS=$(GOOS) $(GOPATH)/bin/dep ensure

update_dep: $(GOPATH)/bin/dep
	GOPATH=$(GOPATH) GOOS=$(GOOS) $(GOPATH)/bin/dep ensure --update

build: build_dep
	GOPATH=$(GOPATH) GOOS=$(GOOS) GOLDFLAGS=$(GOLDFLAGS) go install \
	-ldflags $(GOLDFLAGS) -v github.com/Donders-Institute/hpc-utility/...

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
