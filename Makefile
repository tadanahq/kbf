# kbf — repo gates.
#
# `make check` is the quality gate (AGENTS.md / project-standards.md):
# gofmt, golangci-lint, go test, `kbf lint` over packages/ + examples/
# (dogfood), the conformance suite, schema-freshness, and boundaries.
# Green before any task is marked done.
#
# conformance/ has no fixtures yet (Batch 4): that target degrades to a
# clear skip message and exit 0 rather than fail a fresh clone that hasn't
# reached that batch. Every other target enforces for real: dogfood-lint
# (packages/universal-core, examples/cafe-demo) and boundaries
# (scripts/boundaries.go) both have real content/code to check already.

TOOLS := tools
SCRIPTS := scripts
BIN := bin/kbf

.PHONY: check
check: fmt lint test dogfood-lint conformance schema-freshness boundaries

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
	@if [ -f packages/universal-core/manifest.yaml ] && [ -f examples/cafe-demo/manifest.yaml ]; then \
		./$(BIN) lint packages/universal-core examples/cafe-demo; \
	else \
		echo "dogfood-lint: packages/universal-core or examples/cafe-demo not present yet, skipping"; \
	fi

.PHONY: conformance
conformance:
	@if [ -z "$$(find conformance -type f 2>/dev/null)" ]; then \
		echo "conformance: no fixtures under conformance/ yet, skipping"; \
	else \
		cd $(TOOLS) && go test ./internal/conformance/...; \
	fi

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
