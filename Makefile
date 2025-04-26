# Go format Makefile (Google 스타일)

.PHONY: format

format:
	@echo 'Running gofmt (Google style)...'
	@gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		echo 'Running goimports...'; \
		goimports -w .; \
	else \
		echo 'goimports not found. Install: go install golang.org/x/tools/cmd/goimports@latest'; \
	fi
	@if command -v gofumpt >/dev/null 2>&1; then \
		echo 'Running gofumpt (stricter Google style)...'; \
		gofumpt -w .; \
	else \
		echo 'gofumpt not found. Install: go install mvdan.cc/gofumpt@latest'; \
	fi

.PHONY: init

init:
	@echo 'Installing pre-push git hook...'
	@mkdir -p .git/hooks
	@cat > .git/hooks/pre-push <<'EOF'
#!/bin/sh

# Run tests before pushing

echo 'Running go test...'
go test ./...
if [ $$? -ne 0 ]; then
  echo 'Tests failed. Push aborted.'
  exit 1
fi

echo 'Running go vet...'
go vet ./...
if [ $$? -ne 0 ]; then
  echo 'go vet failed. Push aborted.'
  exit 1
fi

echo 'All tests and static analysis passed. Push allowed.'
EOF
	@chmod +x .git/hooks/pre-push
	@echo 'pre-push hook installed successfully.'
	@echo 'Downloading Go modules...'
	@go mod download
	@echo 'Go modules downloaded.' 