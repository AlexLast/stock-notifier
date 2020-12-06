darwin:
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -a -o build/notifier github.com/alexlast/stock-notifier/cmd/notifier
test:
	go test ./internal/... ./cmd/... -coverprofile cover.out -timeout 20m