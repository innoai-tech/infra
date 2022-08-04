example:
	go run ./cmd/example serve --dump-k8s
	go run ./cmd/example serve

webapp:
	go run ./cmd/example webapp --root ./cmd/example/ui/dist

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