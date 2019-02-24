# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=millipedes
CMD_PATH=./cmd/$(BINARY_NAME)
    
build:
	$(GOBUILD) -o $(BINARY_NAME) $(CMD_PATH)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) $(CMD_PATH)
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	./$(BINARY_NAME)
run-debug:
	./$(BINARY_NAME) -loglevel="debug"
all: test build
build-run: build run
build-run-debug: build run-debug