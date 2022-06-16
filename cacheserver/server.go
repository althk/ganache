package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	pb "github.com/althk/ganache/cacheserver/proto"
	cp "github.com/althk/ganache/cacheserver/strategy"
	csmpb "github.com/althk/ganache/csm/proto"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CachingStrategy interface {
	Get(ctx context.Context, key string) (*pb.CacheValue, bool)
	Set(ctx context.Context, key string, val *pb.CacheValue) (int64, error)
	KeysCount(ctx context.Context) int64
	CurrSize(ctx context.Context) int64 // size of current cache in bytes
}

type server struct {
	pb.UnimplementedCacheServer
	h         *health.Server
	c         CachingStrategy // cache store
	etcdc     *clientv3.Client
	addr      string
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
	resp := &pb.GetResponse{
		Data: &anypb.Any{},
	}
	proto.Merge(resp.Data, v.Data)
	return resp, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*emptypb.Empty, error) {
	go func() {
		atomic.AddUint64(&s.nofReqs, 1)
	}()
	k := s.key(in.Namespace, in.Key)
	ts := timestamppb.Now()
	v := &pb.CacheValue{
		SourceTs: ts,
		Data:     &anypb.Any{},
	}
	proto.Merge(v.Data, in.Data)
	s.c.Set(ctx, k, v)
	go func() {
		atomic.AddUint64(&s.nofSets, 1)
	}()
	go func() {
		rk := fmt.Sprintf("ganache/cache/%d/%s", s.shardNum, k)
		m := &pb.CacheKeyMetadata{
			Source: s.addr,
			Key:    k,
			Value:  &pb.CacheValue{},
		}
		proto.Merge(m.Value, v)
		data, _ := proto.Marshal(m)
		s.etcdc.Put(context.Background(), rk, string(data))
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

func newCacheServer(h *health.Server, n int32, addr string) *server {
	c, err := etcdV3Client(*etcdSpec)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return &server{
		c:        cp.NewLRUCache(*maxCacheBytes),
		h:        h,
		shardNum: n,
		etcdc:    c,
		addr:     addr,
	}
}

func registerCacheServer(s *grpc.Server, cs *server) {
	pb.RegisterCacheServer(s, cs)
	go watchCache(cs, fmt.Sprintf("ganache/cache/%d", cs.shardNum))
	err := registerEndpoint(*csmSpec, *shard, cs.addr)
	if err != nil {
		log.Fatal().Msgf("Registration with Shard mgr failed: %v", err)
	}
	cs.h.SetServingStatus("cs", hpb.HealthCheckResponse_SERVING)
}

func watchCache(cs *server, watchPrefix string) {
	log.Info().Msgf("Setting up watch on %v", watchPrefix)
	rch := cs.etcdc.Watch(context.Background(), watchPrefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, e := range wresp.Events {
			go func(e *clientv3.Event) {
				processCacheEvent(cs, e)
			}(e)
		}

	}
}

func processCacheEvent(cs *server, e *clientv3.Event) {
	d := &pb.CacheKeyMetadata{}
	err := proto.Unmarshal(e.Kv.Value, d)
	if err != nil {
		log.Warn().Msgf("error syncing key %v", e.Kv.Key)
		return
	}
	if d.Source == cs.addr {
		return // ignore self updates
	}
	cv, exists := cs.c.Get(context.Background(), d.Key)
	if !exists || d.Value.SourceTs.AsTime().After(cv.SourceTs.AsTime()) {
		cs.c.Set(context.Background(), d.Key, d.Value)
		go func() {
			atomic.AddInt64(&cs.nofSyncs, 1)
		}()
		return
	}
}

func registerEndpoint(csmSpec string, shard int, addr string) error {
	conn, err := grpc.Dial(csmSpec, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	cli := csmpb.NewShardManagerClient(conn)
	resp, err := cli.RegisterCacheServer(context.Background(), &csmpb.RegisterCacheServerRequest{
		ServerSpec: addr,
		Shard:      int64(shard),
	})
	if err != nil {
		return err
	}
	log.Info().Msgf("cache server registered at path %v", resp.RegisteredPath)
	defer conn.Close()
	return nil
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
