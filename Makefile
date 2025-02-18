export INFRA_CLI_DEBUG = 1

EXAMPLE = go tool example

example:
	$(EXAMPLE) serve -c \
		--server-addr=:8081

example.dump:
	$(EXAMPLE) serve --dump-k8s

webapp:
	$(EXAMPLE) webapp --root ./cmd/example/ui/dist --ver=test

webapp.2:
	$(EXAMPLE) webapp --disable-history-fallback --root ./cmd/example/ui/dist

tidy:
	go mod tidy

test:
	CGO_ENABLED=0 go test -failfast ./...

test.race:
	CGO_ENABLED=1 go test -v -race ./...

fmt:
	go tool gofumpt -l -w .

dep:
	go get -u ./...

gen:
	go tool devtool gen -a ./cmd/example

hey:
	hey -z 5m http://localhost:8081/api/example/v0/orgs