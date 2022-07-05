package server

import (
	"context"
	"time"

	"github.com/althk/ganache/cacheserver/internal/config"
	"github.com/althk/ganache/cacheserver/internal/service"
	"github.com/althk/ganache/cacheserver/internal/strategy"
	csync "github.com/althk/ganache/cacheserver/internal/sync"
	csmpb "github.com/althk/ganache/csm/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func registerWithCSM(cscfg *config.CSConfig) error {
	creds, err := cscfg.TLSConfig.Creds()
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(cscfg.CSMSpec, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	cli := csmpb.NewShardManagerClient(conn)
	resp, err := cli.RegisterCacheServer(context.Background(), &csmpb.RegisterCacheServerRequest{
		ServerSpec: cscfg.Addr,
		Shard:      int64(cscfg.Shard),
	})
	if err != nil {
		return err
	}
	log.Info().Msgf("cache service.CacheServer registered at path %v", resp.RegisteredPath)
	defer conn.Close()
	return nil
}

func etcdV3Client(etcdSpec string) (*clientv3.Client, error) {
	log.Info().Msgf("Connecting to etcd service.CacheServer: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func New(cscfg *config.CSConfig) (*service.CacheServer, error) {
	lru := strategy.NewLRUCache(cscfg.MaxCacheBytes)
	etcdc, err := etcdV3Client(cscfg.ETCDSpec)
	if err != nil {
		return nil, err
	}
	cacheServer, err := service.NewCacheServer(cscfg, lru, etcdc)
	if err != nil {
		return nil, err
	}
	if err = csync.InitWatchAndSync(cacheServer); err != nil {
		return nil, err
	}
	if err = registerWithCSM(cscfg); err != nil {
		return nil, err
	}
	return cacheServer, nil
}
