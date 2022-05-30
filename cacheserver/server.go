package main

import (
	"context"
	"fmt"
	"sync/atomic"

	pb "github.com/althk/ganache/cacheserver/proto"
	cp "github.com/althk/ganache/cacheserver/strategy"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CachingStrategy interface {
	Get(ctx context.Context, key string) (*anypb.Any, bool)
	Set(ctx context.Context, key string, val *anypb.Any) (int64, error)
	KeysCount(ctx context.Context) int64
	CurrSize(ctx context.Context) int64 // size of current cache in bytes
}

type server struct {
	pb.UnimplementedCacheServer
	h         *health.Server
	c         CachingStrategy // cache store
	nofGets   uint64
	nofSets   uint64
	nofMisses uint64
	nofReqs   uint64
	shardNum  int32
	// TODO: add the following features (no particular order):
	// 1. counters by namespaces
	// 2. expiration policy
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	log.Debug().
		Str("m", "get").
		Str("ns", in.Namespace).
		Str("k", in.Key).
		Send()
	go func() {
		atomic.AddUint64(&s.nofGets, 1)
		atomic.AddUint64(&s.nofReqs, 1)
	}()
	v, e := s.c.Get(ctx, s.key(in.Namespace, in.Key))
	if !e {
		atomic.AddUint64(&s.nofMisses, 1)
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	return &pb.GetResponse{Data: v}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	go func() {
		atomic.AddUint64(&s.nofReqs, 1)
	}()
	s.c.Set(ctx, s.key(in.Namespace, in.Key), in.Data)
	go func() {
		atomic.AddUint64(&s.nofSets, 1)
	}()
	return &emptypb.Empty{}, nil
}

func (s *server) Stats(ctx context.Context, _ *emptypb.Empty) (*pb.StatsResponse, error) {
	return &pb.StatsResponse{
		GetReqCount:         s.nofGets,
		SetReqCount:         s.nofSets,
		TotalReqCount:       s.nofReqs,
		TotalCacheSizeBytes: uint64(s.c.CurrSize(ctx)),
		TotalKeysCount:      uint64(s.c.KeysCount(ctx)),
		ShardNumber:         s.shardNum,
	}, nil
}

func (s *server) key(ns, key string) string {
	return fmt.Sprintf("%s%s", ns, key)
}

func newCacheServer(h *health.Server, n int32) *server {
	return &server{
		c:        cp.NewLRUCache(*maxCacheBytes),
		h:        h,
		shardNum: n,
	}
}

func registerCacheServer(s *grpc.Server, cs *server) {
	pb.RegisterCacheServer(s, cs)
	cs.h.SetServingStatus("cs", hpb.HealthCheckResponse_SERVING)
}
