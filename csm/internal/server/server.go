package server

import (
	"os"
	"time"

	"github.com/althk/ganache/csm/internal/service"
	clientv3 "go.etcd.io/etcd/client/v3"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func New(etcdSpec, resolverPrefix string) (*service.CSM, error) {
	etcdc, err := etcdV3Client(etcdSpec)
	if err != nil {
		return nil, err
	}
	return service.NewCSM(etcdc, resolverPrefix)
}

func GetGRPCServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		getServerInterceptorChain(),
	}
}

func getServerInterceptorChain() grpc.ServerOption {
	logger := zerolog.New(os.Stdout)
	return middleware.WithUnaryServerChain(
		tags.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(logger)),
	)
}

func etcdV3Client(etcdSpec string) (*clientv3.Client, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil

}
