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