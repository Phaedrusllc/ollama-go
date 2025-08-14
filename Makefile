GO ?= go

.PHONY: fmt lint vet test e2e examples

fmt:
	@gofmt -s -w .

lint:
	@golangci-lint run ./...

vet:
	@$(GO) vet ./...

test:
	@$(GO) test ./... -v -count=1

e2e:
	@RUN_E2E=1 $(GO) test ./examples -v -count=1

