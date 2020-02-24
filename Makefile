GIT_VER := $(shell git describe --tags)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
export GO111MODULE := on

.PHONY: test binary install clean dist
cmd/proplo/proplo: *.go cmd/proplo/*.go
	cd cmd/proplo && go build -ldflags "-s -w -X main.Version=${GIT_VER}" -gcflags="-trimpath=${PWD}"

install: cmd/proplo/proplo
	install cmd/proplo/proplo ${GOPATH}/bin

test:
	go test -race .
	go test -race ./cmd/proplo

clean:
	rm -f cmd/proplo/proplo
	rm -fr dist/

dist:
	CGO_ENABLED=0 \
		goxz -pv=$(GIT_VER) \
		-build-ldflags="-s -w -X main.Version=${GIT_VER}" \
		-os=darwin,linux -arch=amd64 -d=dist ./cmd/proplo

release:
	ghr -u fujiwara -r proplo -n "$(GIT_VER)" $(GIT_VER) dist/
