package strategy

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/althk/ganache/cacheserver/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ctx      = context.TODO()
	v        = "10byte val"
	se       = len([]byte(v))
	ts       = timestamppb.Now()
	c        *lru
	maxBytes = int64(100)
	ll       *dll
)

func TestLRUNew(t *testing.T) {
	c = NewLRUCache(maxBytes)
	require.NotNil(t, c)
	require.EqualValues(t, 0, c.CurrSize(context.TODO()))
	require.EqualValues(t, 0, c.Count(context.TODO()))

	require.Equal(t, 10, se) // assert pre-condition
}

func TestLRUSetSomeValues(t *testing.T) {
	populateCache(t)
}

func TestLRUSetRemovesOldestItem(t *testing.T) {
	populateCache(t)

	// check LRU impact
	require.True(t, c.cm.Has("key_1"))
	s, err := c.Set(ctx, "key_11", cacheValue(v, ts))
	time.Sleep(time.Millisecond * 1) // required for allowing other goroutines to run

	require.NoError(t, err)
	require.EqualValues(t, 10, c.Count(ctx))
	require.EqualValues(t, se, s)
	require.EqualValues(t, se*10, c.CurrSize(ctx))
	require.False(t, c.cm.Has("key_1")) // oldest key should be dropped
	require.True(t, c.cm.Has("key_11"))
}

func TestLRUGetNonexistentItem(t *testing.T) {
	populateCache(t)
	got, e := c.Get(ctx, "nonexistent_key")
	require.Nil(t, got)
	require.False(t, e)
}
func TestLRUGetExistingItem(t *testing.T) {
	populateCache(t)
	got, e := c.Get(ctx, "key_10")
	require.True(t, e)
	require.EqualValues(t, cacheValue(v, ts), got)
}

func TestDLLUpsertFrontNewElement(t *testing.T) {
	ll = &dll{}
	ll.UpsertFront("key1", false)
	require.EqualValues(t, "key1", ll.head.data)
	require.EqualValues(t, "key1", ll.tail.data)

	ll.UpsertFront("key2", false)
	ll.UpsertFront("key3", false)
	require.EqualValues(t, "key3", ll.head.data)
	require.EqualValues(t, "key1", ll.tail.data)
}

func TestDLLUpsertFrontExistingElement(t *testing.T) {
	ll = &dll{}
	ll.UpsertFront("key1", false)
	ll.UpsertFront("key2", false)
	ll.UpsertFront("key3", false)
	ll.UpsertFront("key1", true)
	require.EqualValues(t, "key1", ll.head.data)
	require.EqualValues(t, "key2", ll.tail.data)
}

func TestDLLRemoveBack(t *testing.T) {
	ll = &dll{}
	ll.UpsertFront("key1", false)
	ll.UpsertFront("key2", false)
	ll.UpsertFront("key3", false)
	k := ll.RemoveBack()
	require.EqualValues(t, "key1", k)
	require.EqualValues(t, "key2", ll.tail.data)
	require.EqualValues(t, "key3", ll.head.data)
}

func populateCache(t *testing.T) {
	c = NewLRUCache(maxBytes)
	for i := 1; i <= 10; i++ {
		s, err := c.Set(ctx, fmt.Sprintf("key_%d", i), cacheValue(v, ts))
		time.Sleep(time.Millisecond * 1) // required for allowing other goroutines to run
		require.NoError(t, err)
		require.EqualValues(t, i, c.Count(ctx))
		require.EqualValues(t, se, s)
		require.EqualValues(t, se*i, c.CurrSize(ctx))
	}
}

func cacheValue(v string, ts *timestamppb.Timestamp) *pb.CacheValue {
	return &pb.CacheValue{
		Data: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
			Value:   []byte(v),
		},
		SourceTs: ts,
	}
}
