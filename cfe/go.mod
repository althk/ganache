module github.com/althk/ganache/cfe

go 1.18

require (
	github.com/althk/ganache/cacheserver v0.0.0-20220706175043-8ffe22299080
	github.com/althk/ganache/utils v0.0.0-20220706175043-8ffe22299080
	github.com/althk/goeasy/grpcutils v0.0.0-20220712184942-d7de7754eb7f
	github.com/rs/zerolog v1.27.0
	github.com/stretchr/testify v1.8.0
	go.etcd.io/etcd/client/v3 v3.5.4
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
)

require go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect

require (
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.3-0.20220203105225-a9a7ef127534 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2 v2.0.0-rc.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.3 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.33.0 // indirect
	go.opentelemetry.io/otel v1.8.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.8.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.8.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.8.0 // indirect
	go.opentelemetry.io/otel/sdk v1.8.0 // indirect
	go.opentelemetry.io/otel/trace v1.8.0 // indirect
	go.opentelemetry.io/proto/otlp v0.18.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/net v0.0.0-20220708220712-1185a9018129 // indirect
	golang.org/x/sys v0.0.0-20220712014510-0a85c31ab51e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220712132514-bdd2acd4974d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/althk/ganache/cacheserver => ../cacheserver

replace github.com/althk/ganache/utils => ../utils

replace github.com/althk/ganache/client => ../client
