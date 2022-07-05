package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/althk/ganache/csm/internal/server"
	pb "github.com/althk/ganache/csm/proto"
	grpcutils "github.com/althk/ganache/utils/grpc"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
var etcdSpec = flag.String("etcd_server", "localhost:2379", "address of etcd server in the form host:port")
var csResolverPrefix = flag.String("cacheserver_resolver_prefix", "ganache/cacheserver", "key prefix for cache service resolver")
var debug = flag.Bool("debug", false, "enable debug logging")
var clientCAPath = flag.String("client_ca_file", "", "Path to CA cert file that can verify client certs")
var tlsCrtPath = flag.String("tls_cert_file", "", "Path to server's TLS cert file")
var tlsKeyPath = flag.String("tls_key_file", "", "Path to server's TLS key file")
var skipTLS = flag.Bool("skip_tls", false, "If server should skip TLS and use insecure creds")

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

	csmServer, err := server.New(*etcdSpec, *csResolverPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new CSM server.")
	}
	tlsCfg := &grpcutils.TLSConfig{
		CertFilePath:     *tlsCrtPath,
		KeyFilePath:      *tlsKeyPath,
		ClientCAFilePath: *clientCAPath,
		SkipTLS:          *skipTLS,
	}
	grpcServerCfg := &grpcutils.GRPCServerConfig{
		TLSConfig:          tlsCfg,
		EnableReflection:   true,
		EnableHealthServer: true,
	}
	s, err := grpcutils.NewGRPCServer(grpcServerCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load grpc server opts.")
	}

	// register CSM server
	pb.RegisterShardManagerServer(s, csmServer)

	log.Info().Msgf("Running shard manager server on %v", lis.Addr().String())
	s.Serve(lis)
}
