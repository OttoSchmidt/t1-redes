SERVER_BIN=pacman_server
CLIENT_BIN=pacman_client
GO=go

.PHONY: all debug client server clean

BUILD_FLAGS?=

all: client server

debug:
	$(eval BUILD_FLAGS += -tags debug)

client:
	$(GO) build $(BUILD_FLAGS) -o $(CLIENT_BIN) client/main.go

server:
	$(GO) build $(BUILD_FLAGS) -o $(SERVER_BIN) server/main.go	

clean:
	rm -f $(SERVER_BIN) $(CLIENT_BIN)
	$(GO) clean