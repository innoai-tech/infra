export INFRA_CLI_DEBUG := "1"
example := "go tool example"
devtool := "go tool devtool"

serve-example:
    {{ example }} serve -c \
    	--server-addr=:8081

dump-k8s-example:
    {{ example }} serve --dump-k8s

webapp:
    {{ example }} webapp --root ./cmd/example/ui/dist --ver=test

webapp-debug:
    {{ example }} webapp --disable-history-fallback --root ./cmd/example/ui/dist

dep:
    go mod tidy

update:
    go get -u ./...

test:
    CGO_ENABLED=0 go test -count=1 -failfast ./...

test-race:
    CGO_ENABLED=1 go test -count=1 -race ./...

fmt:
    {{ devtool }} fmt -l -w .

gen:
    {{ devtool }} gen -a ./cmd/example

hey:
    hey -z 5m http://localhost:8081/api/example/v0/orgs
