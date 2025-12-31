.PHONY: proto generate clean build run

proto:
	@echo "Generating protobuf files..."
	@mkdir -p proto
	protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/test.proto

generate: proto

clean:
	rm -rf proto/*.pb.go

build: generate
	cd .. && go build -o testserver/testserver ./testserver

run: build
	./testserver
