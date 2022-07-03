package cfe

import (
	"fmt"
	"os"
	"strings"
	"time"

	cspb "github.com/althk/ganache/cacheserver/proto"
	"github.com/althk/ganache/cfe/internal/service"
	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	resolverv3 "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func New(etcdSpec, csResolverPrefix string, shardCount int) (*service.CFE, error) {
	r, err := etcdResolver(etcdSpec)
	if err != nil {
		return nil, err
	}
	c := make(map[int]cspb.CacheClient)
	for i := 0; i < shardCount; i++ {
		c[i], err = getCacheCli(r, csResolverPrefix, i)
		if err != nil {
			return nil, err
		}
	}
	return service.NewCFE(shardCount, c)

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

func getCacheCli(r resolver.Builder, cacheResolverPrefix string, shardNum int) (cspb.CacheClient, error) {
	ep := strings.Join([]string{"etcd://", cacheResolverPrefix, fmt.Sprint(shardNum)}, "/")
	conn, err := grpc.Dial(ep, grpc.WithResolvers(r), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := cspb.NewCacheClient(conn)
	return c, nil
}

func etcdResolver(etcdSpec string) (resolver.Builder, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	r, err := resolverv3.NewBuilder(cli)
	if err != nil {
		return nil, err
	}
	return r, nil

}
