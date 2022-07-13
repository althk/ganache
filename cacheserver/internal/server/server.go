package server

import (
	"context"

	"github.com/althk/ganache/cacheserver/internal/config"
	"github.com/althk/ganache/cacheserver/internal/service"
	"github.com/althk/ganache/cacheserver/internal/strategy"
	csync "github.com/althk/ganache/cacheserver/internal/sync"
	csmpb "github.com/althk/ganache/csm/proto"
	etcdutils "github.com/althk/ganache/utils/etcd"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func registerWithCSM(cscfg *config.CSConfig) error {
	opts, err := cscfg.ServerConfig.GetGRPCDialOpts()
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(cscfg.CSMSpec, opts...)
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

func New(cscfg *config.CSConfig) (*service.CacheServer, error) {
	lru := strategy.NewLRUCache(cscfg.MaxCacheBytes)
	etcdc, err := etcdutils.V3Client(cscfg.ETCDSpec)
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
