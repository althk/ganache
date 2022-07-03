package service

import (
	"context"
	"testing"
	"time"

	"github.com/althk/ganache/cacheserver/internal/strategy"
	pb "github.com/althk/ganache/cacheserver/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var cache CachingStrategy
var ctx = context.TODO()
var ts = timestamppb.Now()
var val = "some value"
var cs *CacheServer
var kv *mockkv

func TestGetExistingItem(t *testing.T) {
	cache = strategy.NewLRUCache(100)
	cache.Set(ctx, "ns1key1", cacheValue(val, ts))
	cs = &CacheServer{Cache: cache}

	resp, err := cs.Get(ctx, &pb.GetRequest{
		Namespace: "ns1",
		Key:       "key1",
	})
	require.NoError(t, err)
	require.EqualValues(t, []byte(val), resp.Data.Value)
}

func TestGetNonexistentItem(t *testing.T) {
	cache = strategy.NewLRUCache(100)
	cs = &CacheServer{Cache: cache}

	resp, err := cs.Get(ctx, &pb.GetRequest{
		Namespace: "ns1",
		Key:       "key1",
	})
	require.Nil(t, resp)
	require.NotNil(t, err)
	es := status.Convert(err)
	require.Equal(t, codes.NotFound, es.Code())
}

func TestSet(t *testing.T) {
	cache = strategy.NewLRUCache(100)
	mockEtcd := mockETCD()
	cs = &CacheServer{Cache: cache, Etcd: mockEtcd}

	_, err := cs.Set(ctx, &pb.SetRequest{
		Namespace: "ns1",
		Key:       "key1",
		Data: &anypb.Any{
			Value: []byte(val),
		},
		Global: false,
	})
	time.Sleep(time.Second * 1) // needed to allow service goroutines to run.
	require.Nil(t, err)
	got, e := cache.Get(ctx, "ns1key1")
	require.True(t, e)
	require.EqualValues(t, []byte(val), got.Data.Value)
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

func mockETCD() *clientv3.Client {
	kv = &mockkv{}
	return &clientv3.Client{KV: kv}
}

// mockkv implements clientv3.KV
type mockkv struct {
	mock.Mock
}

func (kv *mockkv) Put(_ context.Context, k, v string, _ ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return nil, nil
}

func (kv *mockkv) Get(_ context.Context, _ string, _ ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return nil, nil
}

func (kv *mockkv) Delete(_ context.Context, _ string, _ ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return nil, nil
}

func (kv *mockkv) Compact(_ context.Context, _ int64, _ ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (kv *mockkv) Do(_ context.Context, _ clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (kv *mockkv) Txn(_ context.Context) clientv3.Txn {
	return nil
}
