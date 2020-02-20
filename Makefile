TARGET=bin

.PHONY: lint
lint:
	./scripts/linter.sh

.PHONY: build
build:
	go build -o ${TARGET}/pulsar cmd/pulsar/main.go
	go build -o ${TARGET}/db cmd/ledger/main.go

.PHONY: run
run: build
	osascript console.scpt

.PHONY: stop
stop:
	./scripts/shutdown.sh

.PHONY: test
test:
	go test -v ./...