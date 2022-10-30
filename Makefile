all: build-protobuf build

build-protobuf:
	protoc --go_out=. *.proto

build:
	go build ./cmd/storage/storage.go

test:
	./storage -d ./my-store -s 64GB
