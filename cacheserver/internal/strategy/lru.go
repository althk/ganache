package strategy

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/althk/dmap"
	pb "github.com/althk/ganache/cacheserver/proto"
)

// lru implements CachingStrategy and provides a LRU cache.
type lru struct {
	cm        dmap.DMap[string, *pb.CacheValue]
	ll        *dll  // doubly linked list
	maxBytes  int64 // total cache size
	currBytes int64 // current cache size
}

func (c *lru) Get(_ context.Context, k string) (*pb.CacheValue, bool) {
	v, e := c.cm.Get(k)
	if e {
		go func() {
			c.ll.UpsertFront(k, true)
		}()
	}
	return v, e
}

func (c *lru) Set(_ context.Context, k string, v *pb.CacheValue) (int64, error) {
	s := int64(len(v.Data.Value))
	e := c.cm.Has(k)
	c.cm.Set(k, v)
	go func() {
		atomic.AddInt64(&c.currBytes, s)
		c.ll.UpsertFront(k, e)
		if atomic.LoadInt64(&c.currBytes) > atomic.LoadInt64(&c.maxBytes) {
			rk := c.ll.RemoveBack()
			rv, _ := c.cm.Get(rk)
			rs := int64(len(rv.Data.Value))
			c.cm.Remove(rk)
			atomic.AddInt64(&c.currBytes, -rs)
		}
	}()
	return s, nil
}

func (c *lru) Count(_ context.Context) int64 {
	return int64(c.cm.Count())
}

func (c *lru) CurrSize(_ context.Context) int64 {
	return atomic.LoadInt64(&c.currBytes)
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
	defer l.l.Unlock()
	nh := &dlln{
		data: k,
		next: l.head,
		prev: nil,
	}
	if l.head != nil {
		l.head.prev = nh
	}
	l.head = nh
	if l.tail == nil {
		l.tail = l.head
	}
}

func (l *dll) RemoveBack() string {
	l.l.Lock()
	defer l.l.Unlock()
	v := l.tail.data
	l.tail = l.tail.prev
	l.tail.next = nil
	return v
}

func (l *dll) MoveToFront(k string) {
	l.l.Lock()
	defer l.l.Unlock()
	if l.head.data == k {
		return // node is already head
	}
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

			// adjust tail
			if n.next == nil {
				l.tail = n.prev
			}
			break
		}
		n = n.next
	}
}

func (l *dll) UpsertFront(k string, exists bool) {
	if exists {
		l.MoveToFront(k)
		return
	}
	l.InsertFront(k)
}
