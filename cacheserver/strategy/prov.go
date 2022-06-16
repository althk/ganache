package strategy

import (
	pb "github.com/althk/ganache/cacheserver/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
)

func NewLRUCache(maxBytes int64) *lru {
	return &lru{
		cm:       cmap.New[*pb.CacheValue](),
		ll:       newDDLL(cmap.SHARD_COUNT),
		maxBytes: maxBytes,
	}

}
