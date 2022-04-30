build:
	go build -o vault/plugins/waypoint cmd/vault-plugin-secrets-waypoint/main.go
test:
	go test -v