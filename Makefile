
build:
	go build -o build/debug/consul-kv-migrate main.go

release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/release/consul-kv-migrate-darwin main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/release/consul-kv-migrate.exe main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/release/consul-kv-migrate-linux main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o build/release/consul-kv-migrate-arm main.go

.PHONY: build release run
