## Shard Manager
Responsible for registering cache servers against respective shard resolvers (etcd name resolver).

#### Notes
After making proto changes, regenerate the stubs by running the following cmd from inside the `proto` directory:
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  *.proto
```