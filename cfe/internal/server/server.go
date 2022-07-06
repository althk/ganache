package server

import (
	"fmt"
	"strings"
	"time"

	cspb "github.com/althk/ganache/cacheserver/proto"
	"github.com/althk/ganache/cfe/internal/service"
	etcdutils "github.com/althk/ganache/utils/etcd"
	grpcutils "github.com/althk/ganache/utils/grpc"
	"github.com/rs/zerolog/log"
	resolverv3 "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

var kacp = keepalive.ClientParameters{
	Time:                8 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             2 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,            // send pings even without active streams
}

func New(tlsCfg *grpcutils.TLSConfig, etcdSpec, csResolverPrefix string, shardCount int) (*service.CFE, error) {
	r, err := etcdResolver(etcdSpec)
	if err != nil {
		return nil, err
	}
	c := make(map[int]cspb.CacheClient)
	creds, _ := tlsCfg.Creds()
	for i := 0; i < shardCount; i++ {
		c[i], err = getCacheCli(creds, r, csResolverPrefix, i)
		if err != nil {
			return nil, err
		}
	}
	return service.NewCFE(shardCount, c)

}

func getCacheCli(creds credentials.TransportCredentials, r resolver.Builder, cacheResolverPrefix string, shardNum int) (cspb.CacheClient, error) {
	ep := strings.Join([]string{"etcd://", cacheResolverPrefix, fmt.Sprint(shardNum)}, "/")
	log.Info().Msgf("Build cacheserver client for %v", ep)
	conn, err := grpc.Dial(
		ep,
		grpc.WithResolvers(r),
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		return nil, err
	}
	c := cspb.NewCacheClient(conn)
	return c, nil
}

func etcdResolver(etcdSpec string) (resolver.Builder, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, _ := etcdutils.V3Client(etcdSpec)
	r, err := resolverv3.NewBuilder(cli)
	if err != nil {
		return nil, err
	}
	return r, nil

}
