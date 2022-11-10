NAME = echo
MODULE = github.com/sa6mwa/echo
VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo 0)
SRC = $(MODULE)/cmd/$(NAME)
GO = CGO_ENABLED=0 go
REPO = echo
TAG = latest

.PHONY: all clean build dependencies test docker

all: clean build

clean:
	rm -f $(NAME)

build: dependencies test $(NAME)

dependencies:
	$(GO) get -v -d ./...

test:
	$(GO) test -cover ./...

$(NAME):
	$(GO) build -v -ldflags '-s -X main.version=$(VERSION)' -o $(NAME) $(SRC)

go.mod:
	go mod init $(MODULE)
	go mod tidy

docker:
	docker build -t $(REPO):$(TAG) .
