package service

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	cspb "github.com/althk/ganache/cacheserver/proto"
	pb "github.com/althk/ganache/cfe/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CFE struct {
	pb.UnimplementedCFEServer
	ShardCount  int
	CacheClient map[int]cspb.CacheClient
}

func (s *CFE) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	log.Debug().
		Str("m", "get").
		Str("ns", in.Namespace).
		Str("k", in.Key).
		Send()
	i := getShardNum(in.Namespace, in.Key, s.ShardCount)
	c := s.CacheClient[i]
	r, err := c.Get(ctx, &cspb.GetRequest{
		Namespace: in.Namespace,
		Key:       in.Key,
	})
	if err != nil {
		es := status.Convert(err)
		switch es.Code() {
		case codes.Unavailable:
			return nil, status.Error(codes.Unavailable, "No cache server available.")
		case codes.NotFound:
			return nil, err
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	resp := &pb.GetResponse{
		Data: &anypb.Any{},
	}
	proto.Merge(resp.Data, r.Data)

	return resp, nil
}

func (s *CFE) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	i := getShardNum(in.Namespace, in.Key, s.ShardCount)
	c := s.CacheClient[i]
	req := &cspb.SetRequest{
		Namespace: in.Namespace,
		Key:       in.Key,
		Data:      &anypb.Any{},
	}
	proto.Merge(req.Data, in.Data)
	r, err := c.Set(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return r, nil
}

func NewCFE(shardCount int, cacheClis map[int]cspb.CacheClient) (*CFE, error) {
	return &CFE{
		ShardCount:  shardCount,
		CacheClient: cacheClis,
	}, nil
}

func getShardNum(ns, key string, mod int) int {
	k := fmt.Sprintf("%s%s", ns, key)
	return int(fnv32(k)) % mod
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
