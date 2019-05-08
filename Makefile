all: tower

Project = github.com/dearcode/doodle/service/debug.Project
GitHash = github.com/dearcode/doodle/service/debug.GitHash
GitTime = github.com/dearcode/doodle/service/debug.GitTime
GitMessage = github.com/dearcode/doodle/service/debug.GitMessage

LDFLAGS += -X "$(Project)=github.com/binlake/tower"
LDFLAGS += -X "$(GitHash)=$(shell git log --pretty=format:'%H' -1)"
LDFLAGS += -X "$(GitTime)=$(shell git log --pretty=format:'%ct' -1)"
LDFLAGS += -X "$(GitMessage)=$(shell git log --pretty=format:'%cn %s %b' -1)"

FILES := $$(find . -name '*.go' | grep -vE 'vendor')
SOURCE_PATH := controllers filters models pb routers 

golint:
	go get github.com/golang/lint/golint

megacheck:
	go get honnef.co/go/tools/cmd/megacheck

lint: golint megacheck
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path"/..."; done;
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;
	go tool vet $(FILES) 2>&1
	megacheck ./...

clean:
	@rm -rf tower

tower:
	go build -o $@ -ldflags '$(LDFLAGS)'
