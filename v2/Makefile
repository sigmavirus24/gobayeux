.PHONY: test bench lint vet

test: vet
	@go test -v -coverprofile=coverage.out --cover . ./extensions/...

coverage.out: test

show-cov: coverage.out
	@go tool cover --func=coverage.out

vet:
	@go vet . ./extensions/...

lint: vet
	@golangci-lint run . ./extensions/...

bench:
	@go test -v --benchmem --bench=. . ./extensions/...
