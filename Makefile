NAME = echo
MODULE = github.com/sa6mwa/echo
SRC = $(MODULE)/cmd/$(NAME)
GO = CGO_ENABLED=0 go
REPO = echo
TAG = latest

.PHONY: all clean build test docker

all: clean build

clean:
	rm -f $(NAME)

build: test $(NAME)

test:
	$(GO) test -cover ./...

$(NAME):
	$(GO) build -v -ldflags=-s -o $(NAME) $(SRC)

go.mod:
	go mod init $(MODULE)
	go mod tidy

docker:
	docker build --network none --build-arg "VERSION=$(shell git describe --tags --long --always --dirty)"  -t $(REPO):$(TAG) .
