#

## Ganache

A simple, distributed, in-memory cache service with TLS auth.

>**NOTE: doing this for practicing building production grade systems, it is still in development.**

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
