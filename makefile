SERVER_BIN=pacman_server
CLIENT_BIN=pacman_client
LIB_MODULES=./lib/crc ./lib/rawSockets
GO=go

.PHONY: all lib client server clean

lib:
	$(GO) build $(LIB_MODULES)

client: lib
	$(GO) build -o $(CLIENT_BIN) client/main.go

server: lib
	$(GO) build -o $(SERVER_BIN) server/main.go	

all: client server

clean:
	rm -f $(SERVER_BIN) $(CLIENT_BIN)
	$(GO) clean $(LIB_MODULES) server/... client/...