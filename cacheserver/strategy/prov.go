package strategy

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewLRUCache(maxBytes int64) *lru {
	return &lru{
		cm:       cmap.New[*anypb.Any](),
		ll:       newDDLL(cmap.SHARD_COUNT),
		maxBytes: maxBytes,
	}

}
