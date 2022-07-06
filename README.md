#

## Ganache

A simple, distributed, in-memory cache service with TLS auth.

>**NOTE: doing this for practicing building production grade systems, it is still in development.**

### Benchmarks

The following benchmark tests (`client/benchmark_test.go`) were run against local CFE server, using `client/client.go` with all Ganache components running on the same host.

```bash
$ cd client && go test -bench=CFE -benchtime=100000x -benchmem
goos: linux
goarch: amd64
pkg: github.com/althk/ganache/client
cpu: Intel(R) Core(TM) i7-1065G7 CPU @ 1.30GHz
BenchmarkCFEGetString-8   	  100000	     82826 ns/op	    5289 B/op	      99 allocs/op
BenchmarkCFEGetInt64-8    	  100000	     83450 ns/op	    5238 B/op	      98 allocs/op
PASS
ok  	github.com/althk/ganache/client	16.660s
```

### Get Started

To try this out locally real quick:

#### Pre-reqs

1. Install etcd
2. Clone the repo.
   1. `git clone https://github.com/althk/ganache`
3. Decide how many shards to use for distributed-ness, default is `1`, i.e, no sharding.

After taking care of the pre-reqs mentioned above:

0. `cd ganache`
1. Open new terminal, start etcd `make run-etcd`
2. Open new terminal, start Cache Shard Manager (CSM) `make run-csm1`
3. Open new terminal, start one or more Cache Servers (at least one for each shard) `make run-cacheserver1`
4. Open new terminal, start Cache Frontend (CFE) `make run-cfe1`

5. Use CFE API to profit (A client package will be added, hopefully soon).

### Overview

![Ganache Service Architecture Overview Diagram](docs/ganache-overview.drawio.svg)
