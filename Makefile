test:
	go test -race ./...

fmt:
	goimports -l -w .

tidy:
	go mod tidy