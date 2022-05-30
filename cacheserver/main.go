package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/reflection"

	hpb "google.golang.org/grpc/health/grpc_health_v1"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
var shard = flag.Int("shard", -1, "shard number for key distribution")
var etcdSpec = flag.String("etcd_server", "", "address of etcd service in the form host:port")
var debug = flag.Bool("debug", false, "enable debug logging")
var maxCacheBytes = flag.Int64("max_cache_bytes", 1000000000, "max size oftotal cache in bytes, defaults to 1GiB")

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal().Msgf("Error listening on port %v: %v", *port, err)
	}

	s := grpc.NewServer()
	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	registerCacheServer(s, newCacheServer(h, int32(*shard)))
	reflection.Register(s)

	log.Info().Msgf("Running cache server on %v", lis.Addr().String())
	s.Serve(lis)
}
