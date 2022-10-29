all: build-protobuf

build-protobuf:
	protoc --go_out=. *.proto

