# High-level targets

NAME=motus

.PHONY: build check run

build: build.local
check: check.imports check.fmt check.lint
run: run.local


## Build targets

TAG=latest
GIT_COMMIT=$(shell git rev-list -1 HEAD --no-abbrev-commit)
IS_PACKR_CMD=$(filter build.packr2,$(MAKECMDGOALS))

.PHONY: build.vendor build.vendor.full build.prepare build.cmd build.local build.packr2

build.vendor:
	GO111MODULE=on go mod vendor

build.vendor.full:
	@rm -fr $(PWD)/vendor
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

build.prepare:
	@mkdir -p $(PWD)/target/
	@rm -f $(PWD)/target/$(NAME)

build.cmd: build.prepare
	GO111MODULE=on $(if $(IS_PACKR_CMD),packr2,go) build -mod=vendor $(BUILD_ARGS) -ldflags "-X github.com/pastequo/motus/cli/dewinter/cmd.GitCommitID=$(GIT_COMMIT) -s -w" -o $(PWD)/target/$(NAME) ./cli/dewinter/main.go

build.local: build.cmd

build.packr2: build.cmd


## Check target

LINT_COMMAND=golangci-lint run $(PWD)/cli/... $(PWD)/motus.go
LINT_RESULT=$(PWD)/lint/result.txt
FILES_LIST=$(PWD)/cli/* $(PWD)/motus.go

.PHONY: check.fmt check.imports check.lint

check.fmt:
	GO111MODULE=on gofmt -s -w $(FILES_LIST)

check.imports:
	GO111MODULE=on goimports -w $(FILES_LIST)

check.lint:
	@rm -fr $(PWD)/lint
	@mkdir -p $(PWD)/lint
	GO111MODULE=on $(LINT_COMMAND) >> $(LINT_RESULT) 2>&1


## Run targets

TXT=turbolol
OK_COUNT=3
AMISS_COUNT=2

.PHONY: run.local run.version

run.local:
	$(PWD)/target/$(NAME) display -t $(TXT) -o $(OK_COUNT) -a $(AMISS_COUNT)

run.version:
	$(PWD)/target/$(NAME) version
