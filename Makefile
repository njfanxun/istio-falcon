export SHELL:=/bin/bash
export SHELLOPTS:=$(if $(SHELLOPTS),$(SHELLOPTS):)pipefail:errexit
BUILD_DATE            := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT            := $(shell git rev-parse HEAD)
GIT_REMOTE            := origin
GIT_BRANCH            := $(shell git rev-parse --symbolic-full-name --verify --quiet --abbrev-ref HEAD)
GIT_TAG               := $(shell git describe --exact-match --tags --abbrev=0  2> /dev/null || echo untagged)
GIT_TREE_STATE        := $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
RELEASE_TAG           := $(shell if [[ "$(GIT_TAG)" =~ ^v[0-9]+\.[0-9]+\.[0-9]+.*$$ ]]; then echo "true"; else echo "false"; fi)
DEV_BRANCH            := $(shell [ $(GIT_BRANCH) = master ] || [ `echo $(GIT_BRANCH) | cut -c -8` = release- ] || [ `echo $(GIT_BRANCH) | cut -c -4` = dev- ] || [ $(RELEASE_TAG) = true ] && echo false || echo true)
SRC                   := $(pwd)
VERSION               := v1.0.0

ifeq ($(RELEASE_TAG),true)
VERSION               := $(GIT_TAG)
endif

override LDFLAGS += \
  -X github.com/njfanxun/istio-falcon/pkg/version.gitVersion=$(VERSION) \
  -X github.com/njfanxun/istio-falcon/pkg/version.buildDate=${BUILD_DATE} \
  -X github.com/njfanxun/istio-falcon/pkg/version.gitCommit=${GIT_COMMIT} \
  -X github.com/njfanxun/istio-falcon/pkg/version.gitTreeState=${GIT_TREE_STATE}

ifneq ($(GIT_TAG),)
	override LDFLAGS += -X github.com/njfanxun/istio-falcon/pkg/version.gitTag=${GIT_TAG}
endif

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

IMG ?= istio-falcon:${VERSION}

.PHONY: docker-build
docker-build: Dockerfile
	docker build -t ${IMG} \
	--build-arg VERSION=$(VERSION) \
	--build-arg BUILD_DATE=$(BUILD_DATE) \
	--build-arg GIT_COMMIT=$(GIT_COMMIT) \
	--build-arg GIT_TREE_STATE=$(GIT_TREE_STATE) \
	--build-arg GIT_TAG=$(GIT_TAG) \
	.
	
.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}