example:
	go run ./cmd/example serve --server-addr=:8081

example.dump:
	go run ./cmd/example serve --dump-k8s

webapp:
	go run ./cmd/example webapp --root ./cmd/example/ui/dist --ver=test

webapp.2:
	go run ./cmd/example webapp --disable-history-fallback --root ./cmd/example/ui/dist

tidy:
	go mod tidy

test:
	CGO_ENABLED=0 go test -failfast ./...

test.race:
	CGO_ENABLED=1 go test -v -race ./...

fmt:
	goimports -l -w .

dep:
	go get -u ./...

gen:
	go run ./internal/cmd/tool gen ./cmd/example

gen.debug:
	go run ./internal/cmd/tool gen --log-level=debug ./cmd/example

hey:
	hey -z 5m http://localhost:8081/api/example/v0/orgs