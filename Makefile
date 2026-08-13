# kbf — repo gates.
#
# `make check` is the quality gate (AGENTS.md / project-standards.md):
# gofmt, golangci-lint, go test, `kbf lint` over packages/ + examples/
# (dogfood), the conformance suite, spec doc-extraction, schema-freshness,
# and boundaries. Green before any task is marked done. Every target
# enforces for real: no empty-pass left anywhere (all content/code this
# gate checks exists).

TOOLS := tools
SCRIPTS := scripts
BIN := bin/kbf

.PHONY: check
check: fmt lint test dogfood-lint conformance spec-examples schema-freshness boundaries

.PHONY: fmt
fmt:
	@for mod in $(TOOLS) $(SCRIPTS); do \
		files="$$(cd $$mod && gofmt -l .)"; \
		if [ -n "$$files" ]; then \
			echo "gofmt would reformat ($$mod):"; echo "$$files"; \
			echo "run: cd $$mod && gofmt -w ."; \
			exit 1; \
		fi; \
	done

.PHONY: lint
lint:
	cd $(TOOLS) && golangci-lint run ./...
	cd $(SCRIPTS) && golangci-lint run ./...

.PHONY: test
test:
	cd $(TOOLS) && go test ./...

$(BIN): $(shell find $(TOOLS) -name '*.go') $(TOOLS)/go.mod $(TOOLS)/go.sum
	@mkdir -p bin
	cd $(TOOLS) && go build -o ../$(BIN) ./cmd/kbf

.PHONY: build
build: $(BIN)

.PHONY: dogfood-lint
dogfood-lint: $(BIN)
	./$(BIN) lint packages/universal-core examples/cafe-demo

.PHONY: conformance
conformance:
	cd $(TOOLS) && go test ./internal/conformance/...

.PHONY: spec-examples
spec-examples:
	cd $(TOOLS) && go test ./internal/docextract/...

.PHONY: schema
schema: $(BIN)
	./$(BIN) schema --out schema

.PHONY: schema-freshness
schema-freshness: $(BIN)
	./$(BIN) schema --check --out schema

.PHONY: boundaries
boundaries:
	cd $(SCRIPTS) && go test ./...
	cd $(SCRIPTS) && go run . --root ..

.PHONY: clean
clean:
	rm -rf bin
