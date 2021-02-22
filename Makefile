.PHONY: proto
proto:
	protoc -I api/registry/ api/registry/registry.proto --go_opt=paths=source_relative --go_out=plugins=grpc:api/registry/

.PHONY: server
server:
	go build -o server -v cmd/server/main.go

.PHONY: client
client:
	go build -o client -v cmd/client/main.go

.PHONY: all
all: proto server client

.PHONY: clean
clean:
	rm server
	rm client
