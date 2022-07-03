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

	"github.com/althk/ganache/csm"
	pb "github.com/althk/ganache/csm/proto"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
var etcdSpec = flag.String("etcd_server", "localhost:2379", "address of etcd server in the form host:port")
var csResolverPrefix = flag.String("cacheserver_resolver_prefix", "ganache/cacheserver", "key prefix for cache service resolver")
var debug = flag.Bool("debug", false, "enable debug logging")

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

	csmServer, err := csm.New(*etcdSpec, *csResolverPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new CSM server.")
	}
	s := grpc.NewServer(csm.GetGRPCServerOpts()...)
	pb.RegisterShardManagerServer(s, csmServer)

	// register other servers
	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	reflection.Register(s)

	log.Info().Msgf("Running shard manager server on %v", lis.Addr().String())
	s.Serve(lis)
}
