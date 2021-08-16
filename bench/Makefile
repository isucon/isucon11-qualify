COMMIT=$(shell cat /etc/REVISION 2>/dev/null || git rev-parse --short HEAD)

GOPRIVATE="github.com/isucon"
GOLDFLAGS=-X main.COMMIT=$(COMMIT)
# race detector
GO_BUILD_RACE:=-race
ifdef RELEASE
	GO_BUILD_RACE:=
endif
GO_BUILD:=go build $(GO_BUILD_RACE) -ldflags "$(GOLDFLAGS)"

GOFILES=$(wildcard *.go **/*.go)
PUBLIC_FILES=$(wildcard ../webapp/public/*.* ../webapp/public/**/*.*)

EXE=./bench
ASSETS=./scenario/assets.go


.PHONY: build
build: $(EXE) ## Build bench. if build release binary, exec `make RELEASE=1`

.PHONY: replace-bench
replace-bench: ## Build bench as ELF and move to ../provisioning/ansible-replace-bench
	@make -B $(EXE) GOOS=linux RELEASE=1
	mv $(EXE) ../provisioning/ansible-replace-bench/

.PHONY: clean
clean: ## Cleanup working directory
	@$(RM) $(EXE)
	@go clean

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: vendor
vendor: $(GOPROTOFILES)
	GOPRIVATE=$(GOPRIVATE) go mod vendor

$(EXE): Makefile go.mod $(GOFILES) $(ASSETS)
	GOOS=$(GOOS) GOPRIVATE=$(GOPRIVATE) $(GO_BUILD) -ldflags "$(GOLDFLAGS)" -o $(EXE) -v github.com/isucon/isucon11-qualify/bench

.PHONY: $(ASSETS)
$(ASSETS): $(PUBLIC_FILES)
	GOOS="" go generate ./gen/assets.go
