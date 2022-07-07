package strategy

import (
	"github.com/althk/dmap"
	pb "github.com/althk/ganache/cacheserver/proto"
)

func NewLRUCache(maxBytes int64) *lru {
	return &lru{
		cm:       dmap.New[string, *pb.CacheValue](26),
		ll:       &dll{},
		maxBytes: maxBytes,
	}

}
