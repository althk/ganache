
clean:
	rm -rf build/*

protoc-cacheserver:
	cd cacheserver/proto && \
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto

protoc-csm:
	cd csm/proto && \
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto

protoc-cfe:
	cd cfe/proto && \
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto

protoc-all: protoc-cacheserver protoc-csm protoc-cfe

build-cacheserver: protoc-cacheserver
	cd cacheserver/cmd/server && \
	go build -o ../../../build/cacheserver .

build-csm: protoc-csm
	cd csm/cmd/server && \
	go build -o ../../../build/csm .

build-cfe: protoc-cfe
	cd cfe && \
	go build -o ../build .

build-all: build-cacheserver build-csm build-cfe

test-cacheserver:
	cd cacheserver && \
	go test -cover -race -v ./...

test-csm:
	cd csm && \
	go test -cover -race -v ./...

test-cfe:
	cd cfe && \
	go test -cover -race -v ./...

cacheserver-all: build-cacheserver test-cacheserver

csm-all: build-csm test-csm

cfe-all: build-cfe test-cfe

etcd:
	cd /tmp && etcd

.PHONY: clean protoc-cacheserver protoc-csm protoc-cfe build-cacheserver build-csm build-cfe test-cacheserver test-csm test-cfe etcd cacheserver-all csm-all cfe-all
