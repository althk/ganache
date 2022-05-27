package main

import (
	"context"
	"sync/atomic"

	pb "github.com/althk/ganache/cacheserver/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedCacheServer
	c         cmap.ConcurrentMap[[]byte]
	nofGets   uint64
	nofSets   uint64
	nofMisses uint64
	// TODO: add counters by namespaces
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	log.Info().
		Str("m", "get").
		Str("ns", in.Namespace).
		Str("k", in.Key).
		Send()
	atomic.AddUint64(&s.nofGets, 1)
	v, e := s.c.Get(in.Key)
	if !e {
		atomic.AddUint64(&s.nofMisses, 1)
		return nil, status.Errorf(codes.NotFound, "Cache miss for key %v", in.Key)
	}
	return &pb.GetResponse{Data: v}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	s.c.Set(in.Key, in.Data)
	atomic.AddUint64(&s.nofSets, 1)
	return &emptypb.Empty{}, nil
}

func newCacheServer() *server {
	return &server{c: cmap.New[[]byte]()}
}

func registerCacheServer(s *grpc.Server, cs pb.CacheServer) {
	pb.RegisterCacheServer(s, cs)
}
