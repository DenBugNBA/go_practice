.PHONY: gen

gen:
	protoc -I=./proto \
	 --go_out=paths=source_relative:./service \
	 --go-grpc_out=paths=source_relative:./service \
	./proto/*.proto

test:
	go test -v -race
