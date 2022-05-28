package main

import (
	"context"
	"sync/atomic"

	pb "github.com/althk/ganache/cacheserver/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedCacheServer
	h         *health.Server
	c         cmap.ConcurrentMap[[]byte]
	nofGets   uint64
	nofSets   uint64
	nofMisses uint64
	nofReqs   uint64
	cacheSize uint64 // total cache size in bytes, only values are counted.
	// TODO: add counters by namespaces
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	log.Info().
		Str("m", "get").
		Str("ns", in.Namespace).
		Str("k", in.Key).
		Send()
	atomic.AddUint64(&s.nofGets, 1)
	atomic.AddUint64(&s.nofReqs, 1)
	v, e := s.c.Get(s.key(in.Namespace, in.Key))
	if !e {
		atomic.AddUint64(&s.nofMisses, 1)
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	return &pb.GetResponse{Data: v}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	s.c.Set(s.key(in.Namespace, in.Key), in.Data)
	atomic.AddUint64(&s.nofSets, 1)
	atomic.AddUint64(&s.cacheSize, uint64(len(in.Data)))
	atomic.AddUint64(&s.nofReqs, 1)
	return &emptypb.Empty{}, nil
}

func (s *server) key(ns, key string) string {
	return ns + key
}

func newCacheServer(h *health.Server) *server {
	return &server{
		c: cmap.New[[]byte](),
		h: h,
	}
}

func registerCacheServer(s *grpc.Server, cs *server) {
	pb.RegisterCacheServer(s, cs)
	cs.h.SetServingStatus("cs", hpb.HealthCheckResponse_SERVING)
}
