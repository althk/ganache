package service

import (
	"context"
	"fmt"

	"github.com/althk/ganache/cacheserver/internal/config"
	pb "github.com/althk/ganache/cacheserver/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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
	Cache    CachingStrategy // cache store
	Etcd     *clientv3.Client
	Addr     string
	shardNum int32
}

func (s *CacheServer) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	v, e := s.Cache.Get(ctx, s.key(in.Namespace, in.Key))
	if !e {
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	resp := &pb.GetResponse{
		Data: v.GetData(),
	}
	return resp, nil
}

func (s *CacheServer) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	k := s.key(in.Namespace, in.Key)
	ts := timestamppb.Now()
	v := &pb.CacheValue{
		SourceTs: ts,
		Data:     in.GetData(),
	}
	s.Cache.Set(ctx, k, v)
	go func() {
		rk := s.fullKeyPath(k)
		m := &pb.CacheKeyMetadata{
			Source: s.Addr,
			Key:    k,
			Value:  v,
		}
		data, _ := proto.Marshal(m)
		s.Etcd.Put(context.Background(), rk, string(data))
	}()
	return &emptypb.Empty{}, nil
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
}

func NewCacheServer(cscfg *config.CSConfig, cs CachingStrategy, etcdc *clientv3.Client) (*CacheServer, error) {
	return &CacheServer{
		Cache:    cs,
		Etcd:     etcdc,
		Addr:     cscfg.Addr,
		shardNum: cscfg.Shard,
	}, nil
}
