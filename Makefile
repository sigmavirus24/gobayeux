test-old:
	go test -v -coverprofile=coverage.out --cover . ./extensions/...
	go vet . ./extensions/...
	golint . ./extensions/...
	golangci-lint run . ./extensions/...
	go tool cover --func=coverage.out

bench-old:
	go test -v --benchmem --bench=. . ./extensions/...

test:
	make -C v2 test

bench:
	make -C v2 bench

show-cov:
	make -C v2 show-cov
