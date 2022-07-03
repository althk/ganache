package sync

import (
	"context"

	"github.com/althk/ganache/cacheserver/internal/service"
	pb "github.com/althk/ganache/cacheserver/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
)

func syncCache(cs *service.CacheServer) error {
	log.Info().Msg("Syncing existing stash of cache")
	resp, err := cs.Etcd.Get(context.Background(), cs.EtcdShardPrefix(), clientv3.WithPrefix())
	if err != nil {
		return err
	}
	log.Info().Msgf("Found %d keys to sync", resp.Count)
	for _, kv := range resp.Kvs {
		d := &pb.CacheKeyMetadata{}
		err := proto.Unmarshal(kv.Value, d)
		if err != nil {
			log.Warn().Msgf("error syncing key %v", kv.Key)
		}
		cs.Cache.Set(context.Background(), d.Key, d.Value)
	}
	log.Info().Msg("Sync successfully completed")
	return nil
}

func watchCache(cs *service.CacheServer) {
	log.Info().Msgf("Setting up watch on %v", cs.EtcdShardPrefix())
	rch := cs.Etcd.Watch(context.Background(), cs.EtcdShardPrefix(), clientv3.WithPrefix())
	for wresp := range rch {
		for _, e := range wresp.Events {
			go func(e *clientv3.Event) {
				processCacheEvent(cs, e)
			}(e)
		}

	}
}

func processCacheEvent(cs *service.CacheServer, e *clientv3.Event) {
	d := &pb.CacheKeyMetadata{}
	err := proto.Unmarshal(e.Kv.Value, d)
	if err != nil {
		log.Warn().Msgf("error syncing key %v", e.Kv.Key)
		return
	}
	if d.Source == cs.Addr {
		return // ignore self updates
	}
	cs.Sync(d.Key, d.Value)
}

func InitWatchAndSync(cs *service.CacheServer) error {
	go watchCache(cs)
	return syncCache(cs)
}
