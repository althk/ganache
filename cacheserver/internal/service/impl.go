package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/althk/ganache/cacheserver/config"
	pb "github.com/althk/ganache/cacheserver/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const etcdCachePrefix = "ganache/cache"

type CachingStrategy interface {
	Get(ctx context.Context, key string) (*pb.CacheValue, bool)
	Set(ctx context.Context, key string, val *pb.CacheValue) (int64, error)
	Count(ctx context.Context) int64
	CurrSize(ctx context.Context) int64 // size of current cache in bytes
}

type CacheServer struct {
	pb.UnimplementedCacheServer
	Cache     CachingStrategy // cache store
	Etcd      *clientv3.Client
	Addr      string
	nofGets   uint64
	nofSets   uint64
	nofMisses uint64
	nofReqs   uint64
	nofSyncs  int64
	shardNum  int32
	// TODO: add the following features (no particular order):
	// 1. counters by namespaces
	// 2. expiration policy
}

func (s *CacheServer) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	go func() {
		atomic.AddUint64(&s.nofGets, 1)
		atomic.AddUint64(&s.nofReqs, 1)
	}()
	v, e := s.Cache.Get(ctx, s.key(in.Namespace, in.Key))
	if !e {
		atomic.AddUint64(&s.nofMisses, 1)
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	resp := &pb.GetResponse{
		Data: &anypb.Any{},
	}
	proto.Merge(resp.Data, v.Data)
	return resp, nil
}

func (s *CacheServer) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	k := s.key(in.Namespace, in.Key)
	ts := timestamppb.Now()
	v := &pb.CacheValue{
		SourceTs: ts,
		Data:     &anypb.Any{},
	}
	proto.Merge(v.Data, in.Data)
	s.Cache.Set(ctx, k, v)
	go func() {
		atomic.AddUint64(&s.nofReqs, 1)
		atomic.AddUint64(&s.nofSets, 1)
	}()
	go func() {
		rk := s.fullKeyPath(k)
		m := &pb.CacheKeyMetadata{
			Source: s.Addr,
			Key:    k,
			Value:  &pb.CacheValue{},
		}
		proto.Merge(m.Value, v)
		data, _ := proto.Marshal(m)
		s.Etcd.Put(context.Background(), rk, string(data))
	}()
	return &emptypb.Empty{}, nil
}

func (s *CacheServer) Stats(ctx context.Context, _ *emptypb.Empty) (*pb.StatsResponse, error) {
	return &pb.StatsResponse{
		GetReqCount:         s.nofGets,
		SetReqCount:         s.nofSets,
		TotalReqCount:       s.nofReqs,
		TotalCacheSizeBytes: uint64(s.Cache.CurrSize(ctx)),
		TotalKeysCount:      uint64(s.Cache.Count(ctx)),
		ShardNumber:         s.shardNum,
	}, nil
}

func (s *CacheServer) key(ns, key string) string {
	return fmt.Sprintf("%s%s", ns, key)
}

func (s *CacheServer) EtcdShardPrefix() string {
	return fmt.Sprintf("%s/%d", etcdCachePrefix, s.shardNum)
}

func (s *CacheServer) fullKeyPath(k string) string {
	return fmt.Sprintf("%s/%s", s.EtcdShardPrefix(), k)
}

func (s *CacheServer) Sync(k string, v *pb.CacheValue) {
	curr, exists := s.Cache.Get(context.Background(), k)
	if exists && curr.SourceTs.AsTime().After(v.SourceTs.AsTime()) {
		return // local cache has newer value
	}
	s.Cache.Set(context.Background(), k, v)
	go func() {
		atomic.AddInt64(&s.nofSyncs, 1)
	}()
}

func NewCacheServer(cscfg *config.CSConfig, cs CachingStrategy, etcdc *clientv3.Client) (*CacheServer, error) {
	return &CacheServer{
		Cache:    cs,
		Etcd:     etcdc,
		Addr:     cscfg.Addr,
		shardNum: cscfg.Shard,
	}, nil
}
