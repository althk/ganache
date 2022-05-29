package main

import (
	"context"
	"fmt"
	"sync/atomic"

	pb "github.com/althk/ganache/cacheserver/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedCacheServer
	h         *health.Server
	c         cmap.ConcurrentMap[interface{}] // cache store
	nofGets   uint64
	nofSets   uint64
	nofMisses uint64
	nofReqs   uint64
	cacheSize uint64 // total cache size in bytes, only values are counted.
	shardNum  int32
	// TODO: add the following features (no particular order):
	// 1. counters by namespaces
	// 2. add max cache size
	// 3. eviction policy (LRU and expiry)
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
	v, e := s.c.Get(s.key(in.Namespace, in.Key))
	if !e {
		atomic.AddUint64(&s.nofMisses, 1)
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	return &pb.GetResponse{Data: v.(*anypb.Any)}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	go func() {
		atomic.AddUint64(&s.nofReqs, 1)
	}()
	s.c.Set(s.key(in.Namespace, in.Key), in.Data)
	go func() {
		atomic.AddUint64(&s.nofSets, 1)
		atomic.AddUint64(&s.cacheSize, uint64(len(in.Data.Value)))
	}()
	return &emptypb.Empty{}, nil
}

func (s *server) Stats(ctx context.Context, _ *emptypb.Empty) (*pb.StatsResponse, error) {
	return &pb.StatsResponse{
		GetReqCount:         s.nofGets,
		SetReqCount:         s.nofSets,
		TotalReqCount:       s.nofReqs,
		TotalCacheSizeBytes: s.cacheSize,
		TotalKeysCount:      uint64(s.c.Count()),
		ShardNumber:         s.shardNum,
	}, nil
}

func (s *server) key(ns, key string) string {
	return fmt.Sprintf("%s%s", ns, key)
}

func newCacheServer(h *health.Server, n int32) *server {
	return &server{
		c:        cmap.New[interface{}](),
		h:        h,
		shardNum: n,
	}
}

func registerCacheServer(s *grpc.Server, cs *server) {
	pb.RegisterCacheServer(s, cs)
	cs.h.SetServingStatus("cs", hpb.HealthCheckResponse_SERVING)
}
