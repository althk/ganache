SHELL=/bin/bash

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
	cd cfe/cmd/server && \
	go build -o ../../../build/cfe .

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

newcert-ca1:
	openssl req -x509 -sha256 -days 30 \
	-nodes -newkey rsa:2048 \
	-subj "/O=althk/OU=ganache/CN=ca1" \
	-keyout certs/testca.key -out certs/testca.crt

newcert-cacheserver1:
	openssl req -newkey rsa:4096 -nodes \
	-keyout certs/cs1.key -out certs/cs1.csr \
	-subj "/O=althk/OU=ganache/CN=cs1"
	openssl x509 -req -in certs/cs1.csr -days 30 \
	-CA certs/testca.crt -CAkey certs/testca.key \
	-CAcreateserial -out certs/cs1.crt \
	-extfile <(printf "subjectAltName=DNS:localhost,DNS:ganache/cacheserver/0,IP:0.0.0.0,IP:127.0.0.1")

newcert-csm1:
	openssl req -newkey rsa:4096 -nodes \
	-keyout certs/csm1.key -out certs/csm1.csr \
	-subj "/O=althk/OU=ganache/CN=csm1"
	openssl x509 -req -in certs/csm1.csr -days 30 \
	-CA certs/testca.crt -CAkey certs/testca.key \
	-CAcreateserial -out certs/csm1.crt \
	-extfile <(printf "subjectAltName=DNS:localhost,IP:0.0.0.0,IP:127.0.0.1")

newcert-cfe1:
	openssl req -newkey rsa:4096 -nodes \
	-keyout certs/cfe1.key -out certs/cfe1.csr \
	-subj "/O=althk/OU=ganache/CN=cfe1"
	openssl x509 -req -in certs/cfe1.csr -days 30 \
	-CA certs/testca.crt -CAkey certs/testca.key \
	-CAcreateserial -out certs/cfe1.crt \
	-extfile <(printf "subjectAltName=DNS:localhost,IP:0.0.0.0,IP:127.0.0.1")

cleancerts:
	rm -f certs/*

gencerts: cleancerts newcert-ca1 newcert-csm1 newcert-cacheserver1 newcert-cfe1

run-etcd:
	cd /tmp && etcd

run-csm1:
	cd csm && \
	go run cmd/server/main.go -port 41443 -client_ca_file ../certs/testca.crt -tls_cert_file ../certs/csm1.crt -tls_key_file ../certs/csm1.key

run-cacheserver1:
	cd cacheserver && \
	go run cmd/server/main.go -debug -port 44443 -csm_server localhost:41443 -root_ca_file ../certs/testca.crt -client_ca_file ../certs/testca.crt -tls_cert_file ../certs/cs1.crt -tls_key_file ../certs/cs1.key

run-cfe1:
	cd cfe && \
	go run cmd/server/main.go -debug -port 40001 -root_ca_file ../certs/testca.crt -tls_cert_file ../certs/cfe1.crt -tls_key_file ../certs/cfe1.key

.PHONY: clean protoc-cacheserver protoc-csm protoc-cfe build-cacheserver build-csm build-cfe test-cacheserver test-csm test-cfe run-etcd cacheserver-all csm-all cfe-all newcert-ca1 newcert-cacheserver1 newcert-csm1 newcert-cfe1 gencerts run-cacheserver1 run-csm1 run-cfe1
