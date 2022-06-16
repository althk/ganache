module github.com/althk/ganache/cacheserver

go 1.18

require (
	github.com/althk/ganache/csm v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.26.1
	go.etcd.io/etcd/client/v3 v3.5.4
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
)

require (
	github.com/golang/protobuf v1.5.2
	github.com/orcaman/concurrent-map/v2 v2.0.0
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220527130721-00d5c0f3be58 // indirect
)

replace github.com/althk/ganache/csm => ../csm
