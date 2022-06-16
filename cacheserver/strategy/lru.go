package strategy

import (
	"context"
	"sync"
	"sync/atomic"

	pb "github.com/althk/ganache/cacheserver/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// lru implements CachingStrategy and provides a LRU cache.
type lru struct {
	cm        cmap.ConcurrentMap[*pb.CacheValue]
	ll        *ddll // distributed doubly linked list
	maxBytes  int64 // total cache size
	currBytes int64 // current cache size
}

func (c *lru) Get(_ context.Context, k string) (*pb.CacheValue, bool) {
	v, e := c.cm.Get(k)
	go func() {
		if e {
			c.ll.UpsertFront(k, true)
		}
	}()
	return v, e
}

func (c *lru) Set(_ context.Context, k string, v *pb.CacheValue) (int64, error) {
	s := int64(len(v.Data.Value))
	e := c.cm.Has(k)
	c.cm.Set(k, v)
	go func() {
		atomic.AddInt64(&c.currBytes, s)
		c.ll.UpsertFront(k, e)
		if c.currBytes > c.maxBytes {
			c.cm.Remove(c.ll.RemoveBack(k))
		}
	}()
	return s, nil
}

func (c *lru) KeysCount(_ context.Context) int64 {
	return int64(c.cm.Count())
}

func (c *lru) CurrSize(_ context.Context) int64 {
	return c.currBytes
}

// lln represents a node in a doubly linked list.
// NOTE: this is a 'hardcoded' implementation for our LRU
// cache and as such, the data it holds is the key of the
// cached item as a string.
type dlln struct {
	data string
	next *dlln // next node pointer
	prev *dlln // prev node pointer
}

// dll represents a doubly linked list.
// We are not using container/list type here because
// the remove operation is not optimal for our LRU use case.
// We need a remove that always simply drops the last element,
// i.e., there is no need to find the element and then delete.
// NOTE: this is a 'hardcoded' impl for the LRU cache, as such
// the impl has only partial functionality of a normal doubly
// linked list.
type dll struct {
	head *dlln
	tail *dlln
	l    sync.Mutex // don't need a rwmutex as we don't read from this ll
}

func (l *dll) InsertFront(k string) {
	l.l.Lock()
	nh := &dlln{
		data: k,
		next: l.head,
		prev: nil,
	}
	if l.head != nil {
		l.head.prev = nh
	}
	l.l.Unlock()
}

func (l *dll) RemoveBack() string {
	l.l.Lock()
	v := l.tail.data
	l.tail = l.tail.prev
	l.tail.next = nil
	l.l.Unlock()
	return v
}

func (l *dll) MoveToFront(k string) {
	l.l.Lock()
	n := l.head
	for n != nil {
		if n.data == k {
			n.prev.next = n.next
			t := l.head
			l.head = &dlln{
				data: n.data,
				next: t,
				prev: nil,
			}
			t.prev = l.head
			break
		}
		n = n.next
	}
	l.l.Unlock()
}

type ddll struct {
	shards     []*dll // distributed/sharded doubly linked list
	shardCount int32
}

func newDDLL(shardCount int) *ddll {
	ddll := &ddll{
		shards:     make([]*dll, shardCount),
		shardCount: int32(shardCount),
	}
	for i := 0; i < shardCount; i++ {
		ddll.shards[i] = &dll{}
	}
	return ddll
}

func (l *ddll) GetShard(k string) *dll {
	return l.shards[uint(fnv32(k))%uint(l.shardCount)]
}

func (l *ddll) UpsertFront(k string, exists bool) {
	shard := l.GetShard(k)
	if exists {
		shard.MoveToFront(k)
		return
	}
	shard.InsertFront(k)
}

func (l *ddll) RemoveBack(k string) string {
	shard := l.GetShard(k)
	return shard.RemoveBack()
}

func (l *ddll) InsertFront(k string) {
	shard := l.GetShard(k)
	shard.InsertFront(k)
}

// fnv32 computes and returns a hash for the given key.
// lifted as-is from
// https://github.com/orcaman/concurrent-map/blob/v2.0.0/concurrent_map.go
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
