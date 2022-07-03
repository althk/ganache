package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/althk/ganache/cfe"
	pb "github.com/althk/ganache/cfe/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
var etcdSpec = flag.String("etcd_server", "localhost:2379", "address of etcd server in the form host:port")
var csResolverPrefix = flag.String("cacheserver_resolver_prefix", "ganache/cacheserver", "key prefix for cache service resolver")
var debug = flag.Bool("debug", false, "enable debug logging")
var shards = flag.Int("shards", 1, "number of shards to use for distribution")

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
	cfeServer, err := cfe.New(*etcdSpec, *csResolverPrefix, *shards)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create CFE server.")
	}
	s := grpc.NewServer()
	pb.RegisterCFEServer(s, cfeServer)

	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	reflection.Register(s)

	log.Info().Msgf("Starting CFE on address %v", lis.Addr().String())
	s.Serve(lis)
}
