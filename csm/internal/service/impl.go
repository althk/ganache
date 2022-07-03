package service

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/althk/ganache/csm/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// csm implements ShardManagerServer service
type CSM struct {
	Etcd             *clientv3.Client
	CSResolverPrefix string
	pb.UnimplementedShardManagerServer
}

func (s *CSM) RegisterCacheServer(ctx context.Context, in *pb.RegisterCacheServerRequest) (*pb.RegisterCacheServerResponse, error) {
	err := s.Etcd.Sync(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Error connecting to etcd server: %v", err)
	}

	shardPrefix := strings.Join([]string{s.CSResolverPrefix, fmt.Sprint(in.Shard)}, "/")
	em, err := endpoints.NewManager(s.Etcd, shardPrefix)
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

func NewCSM(etcdc *clientv3.Client, resolverPrefix string) (*CSM, error) {
	return &CSM{
		Etcd:             etcdc,
		CSResolverPrefix: resolverPrefix,
	}, nil
}
