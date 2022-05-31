package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/althk/ganache/csm/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// csm implements ShardManagerServer service
type csm struct {
	etcdSpec         string
	csResolverPrefix string
	pb.UnimplementedShardManagerServer
}

func (s *csm) RegisterCacheServer(ctx context.Context, in *pb.RegisterCacheServerRequest) (*pb.RegisterCacheServerResponse, error) {
	cli, err := etcdV3Client(s.etcdSpec)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error connecting to etcd server: %v", err)
	}
	defer cli.Close()

	shardPrefix := strings.Join([]string{s.csResolverPrefix, fmt.Sprint(in.Shard)}, "/")
	em, err := endpoints.NewManager(cli, shardPrefix)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error getting endpoints mgr: %v", err)
	}

	log.Info().
		Str("cs_spec", in.ServerSpec).
		Str("key", shardPrefix).
		Msgf("Registering new cache server %v", in.ServerSpec)

	epKey := strings.Join([]string{shardPrefix, in.ServerSpec}, "/")
	err = em.AddEndpoint(ctx, epKey, endpoints.Endpoint{Addr: in.ServerSpec})
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	log.Info().
		Str("epKey", epKey).
		Msg("Cache server endpoint resolver registration successful")
	return &pb.RegisterCacheServerResponse{RegisteredPath: epKey}, nil
}

func etcdV3Client(etcdSpec string) (*clientv3.Client, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil

}
