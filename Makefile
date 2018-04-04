NAME=location
IMAGE=docker.jw4.us/$(NAME)

ifeq ($(BUILD_VERSION),)
	BUILD_VERSION=$(shell git describe --dirty --first-parent --always --tags)
endif

all: image

clean:
	-rm ./$(NAME)
	go clean ./...

image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-s -w" -o $(NAME) .
	docker build -t $(IMAGE):latest -t $(IMAGE):$(BUILD_VERSION) .

push: clean image
	docker push $(IMAGE):$(BUILD_VERSION)
	docker push $(IMAGE):latest

