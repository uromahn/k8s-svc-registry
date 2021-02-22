.PHONY: proto
proto:
	protoc -I api/registry/ api/registry/registry.proto --go_opt=paths=source_relative --go_out=plugins=grpc:api/registry/

server:
	go build -o server cmd/server/main.go

client:
	go build -o client cmd/client/main.go

.PHONY: all
all: proto server client

.PHONY: clean
clean:
	rm server
	rm client
