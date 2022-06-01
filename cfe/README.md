## CFE (Cache Front End)
CFE is the service that handles client requests. It is the 'frontend' for the Cache servers. It sends requests to the correct shard based on the namespace and key.

CFE uses etcd resolver (which is maintained by shard manager service) to get to the correct cache shard.

#### Notes
After making proto changes, regenerate the stubs by running the following cmd from inside the `proto` directory:
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  *.proto
```