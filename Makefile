.PHONY: test fuzz lint gen-proto

GOLANGCI_LINT_VERSION := v2.12.2
GOLANGCI_LINT_BIN     := bin/golangci-lint

$(GOLANGCI_LINT_BIN):
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) to bin/..."
	@mkdir -p bin
	@GOBIN=$(CURDIR)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT_BIN)
	@actual=$$($(GOLANGCI_LINT_BIN) --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	want=$(GOLANGCI_LINT_VERSION:v%=%); \
	if [ "$$actual" != "$$want" ]; then \
		echo "golangci-lint version mismatch: have $$actual, want $$want, reinstalling..."; \
		rm -f $(GOLANGCI_LINT_BIN); \
		$(MAKE) $(GOLANGCI_LINT_BIN); \
	fi

lint: golangci-lint
	$(GOLANGCI_LINT_BIN) run ./...

test:
	go test -race ./...

fuzz:
	go test -list 'Fuzz.*' . | grep '^Fuzz' | while read f; do \
		go test -fuzz="^$${f}$$" -fuzztime=5s .; \
	done

gen-proto:
	./internal/test/generate.sh
