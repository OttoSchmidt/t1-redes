SERVER_BIN=pacman_server
CLIENT_BIN=pacman_client
GO=go

.PHONY: all debug vendor client server test dist clean

TEST_FILES=lib/rawSockets/tests/*.go

BUILD_FLAGS?=

all: client server

debug: BUILD_FLAGS += -tags debug
debug: all

vendor: BUILD_FLAGS += -mod=vendor
vendor: all

client:
	$(GO) build $(BUILD_FLAGS) -o $(CLIENT_BIN) client/main.go

server:
	$(GO) build $(BUILD_FLAGS) -o $(SERVER_BIN) server/main.go	

test:
	$(GO) test -v $(TEST_FILES)

dist:
	mkdir -p dist
	cp -r client/ lib/ server/ makefile go.mod go.sum dist/
	tar -czvf pacman-redes.tar.gz dist/
	rm -rf dist

clean:
	rm -f $(SERVER_BIN) $(CLIENT_BIN)
	$(GO) clean