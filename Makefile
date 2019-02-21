PREFIX ?= "/opt/cluster"

VERSION ?= "master"

all: build

$(GOPATH)/bin/dep:
	mkdir -p $(GOPATH)/bin
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | GOPATH=$(GOPATH) GOOS=linux sh

build_dep: $(GOPATH)/bin/dep
	GOPATH=$(GOPATH) GOOS=linux $(GOPATH)/bin/dep ensure

update_dep: $(GOPATH)/bin/dep
	GOPATH=$(GOPATH) GOOS=linux $(GOPATH)/bin/dep ensure --update

build: build_dep
	GOPATH=$(GOPATH) GOOS=linux go install github.com/Donders-Institute/hpc-cluster-tools/...

doc:
	@GOPATH=$(GOPATH) GOOS=linux godoc -http=:6060

test: build_dep
	@GOPATH=$(GOPATH) GOOS=linux GOCACHE=off go test -v github.com/Donders-Institute/hpc-cluster-tools/test/...

install: build
	@install -D $(GOPATH)/bin/* $(PREFIX)/bin

release:
	VERSION=$(VERSION) rpmbuild --undefine=_disable_source_fetch -bb build/rpm/centos7.spec

clean:
	@rm -rf $(GOPATH)/bin/cluster-*
	@rm -rf $(GOPATH)/pkg/*/Donders-Institute/hpc-cluster-tools
	@rm -rf $(GOPATH)/pkg/*/Donders-Institute/hpc-cluster-tools
