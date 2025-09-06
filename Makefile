.PHONY: mocks test fieldalignment clean

# Generate mocks
mocks:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=repository/repository.go -destination=mocks/repository_mock.go -package=mocks

# Run tests
test: mocks
	go test -race ./...

# Check struct field alignment (excluding mocks and test files)
fieldalignment:
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.14.0
	@for dir in config constant model pkg pubsub repository; do \
		fieldalignment ./$$dir 2>/dev/null || true; \
	done | grep -v "_test.go:" || true

# Clean everything
clean:
	go clean
	find . -name "*_mock*.go" -delete
	rm -f coverage.out
