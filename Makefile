server:
	go build -o registry-server cmd/registry-server/server.go

client:
	go build -o registry-client cmd/registry-client/client.go

all: server client

clean:
	rm registry-server
	rm registry-client
