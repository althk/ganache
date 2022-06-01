package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	cspb "github.com/althk/ganache/cacheserver/proto"
	pb "github.com/althk/ganache/cfe/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	resolverv3 "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pb.UnimplementedCFEServer
	h                   *health.Server
	etcdSpec            string
	cacheResolverPrefix string
	shardCount          int
	resolver            resolver.Builder
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	log.Debug().
		Str("m", "get").
		Str("ns", in.Namespace).
		Str("k", in.Key).
		Send()
	i := getShardNum(in.Namespace, in.Key, s.shardCount)
	c, err := s.getCacheCli(ctx, i)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	r, err := c.Get(ctx, &cspb.GetRequest{
		Namespace: in.Namespace,
		Key:       in.Key,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetResponse{Data: r.Data}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	i := getShardNum(in.Namespace, in.Key, s.shardCount)
	c, err := s.getCacheCli(ctx, i)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	r, err := c.Set(ctx, &cspb.SetRequest{
		Namespace: in.Namespace,
		Key:       in.Key,
		Data:      in.Data,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return r, nil
}

func getShardNum(ns, key string, mod int) int {
	k := fmt.Sprintf("%s%s", ns, key)
	return int(fnv32(k)) % mod
}

func (s *server) getCacheCli(ctx context.Context, shardNum int) (cspb.CacheClient, error) {
	ep := strings.Join([]string{"etcd://", s.cacheResolverPrefix, fmt.Sprint(shardNum)}, "/")
	conn, err := grpc.Dial(ep, grpc.WithResolvers(s.resolver), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := cspb.NewCacheClient(conn)
	return c, nil
}

func etcdResolver(etcdSpec string) (resolver.Builder, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	r, err := resolverv3.NewBuilder(cli)
	if err != nil {
		return nil, err
	}
	return r, nil

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
