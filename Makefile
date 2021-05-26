.PHONY: proto
proto:
	protoc -I api/registry/ api/registry/registry.proto --go_opt=paths=source_relative --go_out=plugins=grpc:api/registry/

.PHONY: server
server:
	go build -o server -v cmd/server/main.go

.PHONY: client
client:
	go build -o client -v cmd/client/main.go

.PHONY: testall
testall: testserver testclient

.PHONY: testserver
testserver:
	go test -v -timeout 30s -run TestRegister github.com/uromahn/k8s-svc-registry/pkg/registry/server

.PHONY: testclient
testclient:
	go test -v -timeout 30s -run TestReadConfig github.com/uromahn/k8s-svc-registry/cmd/client/config

.PHONY: all
all: proto server client

.PHONY: clean
clean:
	rm server
	rm client
