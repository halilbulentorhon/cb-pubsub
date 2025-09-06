.PHONY: mocks clean-mocks

mocks:
	mockgen -source=repository/repository.go -destination=mocks/repository_mock.go -package=mocks

clean-mocks:
	find . -name "*_mock*.go" -delete
