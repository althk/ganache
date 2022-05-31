## Cache server
Provides an in-memory cache server.
Currently defaults to an LRU cache with a default max size of 1GiB per instance.

#### Notes
After making proto changes, regenerate the stubs by running the following cmd from inside the `proto` directory:
```sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  *.proto
```