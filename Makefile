ORG     := $(shell basename $(realpath ..))
PKGS    := $(shell go list ./... | grep -v /vendor/)
TARGETS := ./plugins/...
PREFIX  := quay.io/roboll

build:
	go build ${TARGETS}
.PHONY: build

generate:
	go generate ${PKGS}
.PHONY: generate

fmt:
	go fmt ${PKGS}
.PHONY: fmt

check:
	go vet ${PKGS}
.PHONY: check

test:
	go test -v ${PKGS} -cover -race -p=1
.PHONY: test

cross:
	gox -os '!freebsd' -arch '!arm' -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" ${TARGETS}
.PHONY: cross

pristine: generate fmt
	git ls-files --modified --deleted --untracked | diff /dev/null -
.PHONY: pristine

release: pristine cross
	@ghr -b ${BODY} -t ${GITHUB_TOKEN} -u ${ORG} -replace ${TAG} dist
.PHONY: release

tools:
	go get -u github.com/roboll/ghr github.com/mitchellh/gox
.PHONY: tools

images:
	docker build -t ${PREFIX}/vault:${TAG} images/vault
.PHONY: images

push: images
	docker push ${PREFIX}/vault:${TAG}
.PHONY: push

TAG  = $(shell git describe --tags --abbrev=0 HEAD)
LAST = $(shell git describe --tags --abbrev=0 HEAD^)
BODY = "`git log ${LAST}..HEAD --oneline --decorate` `printf '\n\#\#\# [Build Info](${BUILD_URL})'`"
