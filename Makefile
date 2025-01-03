# Change these variables as necessary.
main_package_path = ./

livemonitor = *.go

.PHONY: live
live:
	ls -1d ${livemonitor} | entr -r -s "make $(LIVETARGET)"

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)"


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test/all
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1020,-ST1021,-ST1022,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run quick tests
.PHONY: test
test:
	go test -v -short -race -buildvcs ./...

## test/live: make test every time a file changes
.PHONY: test/live
test/live: ${tailwindcss}
	make LIVETARGET=test live

## test/all: run all tests
.PHONY: test/all
test/all:
	go test -tags integration -v -race -buildvcs ./...

## test/all/live: make testall every time a file changes
.PHONY: test/all/live
test/all/live: ${tailwindcss}
	make LIVETARGET=test/all live

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -tags integration -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fmt ./...


# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## push: push changes to the remote Git repository
.PHONY: push
push: confirm audit no-dirty
	git push
