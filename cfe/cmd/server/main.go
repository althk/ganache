package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/althk/ganache/cfe/internal/server"
	pb "github.com/althk/ganache/cfe/proto"
	"github.com/althk/goeasy/grpcutils"
)

var (
	port             = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
	etcdSpec         = flag.String("etcd_server", "", "address of etcd server in the form host:port")
	csResolverPrefix = flag.String("cacheserver_resolver_prefix", "ganache/cacheserver", "key prefix for cache service resolver")
	debug            = flag.Bool("debug", false, "enable debug logging")
	shards           = flag.Int("shards", 1, "number of shards to use for distribution")
	clientCAPath     = flag.String("client_ca_file", "", "Path to CA cert file that can verify client certs")
	rootCAPath       = flag.String("root_ca_file", "", "Path to CA cert file that can verify server/peer certs")
	tlsCrtPath       = flag.String("tls_cert_file", "", "Path to server's TLS cert file")
	tlsKeyPath       = flag.String("tls_key_file", "", "Path to server's TLS key file")
	skipTLS          = flag.Bool("skip_tls", false, "If server should skip TLS and use insecure creds")
)

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatal().Msgf("Error listening on port %v: %v", *port, err)
	}

	tp, err := grpcutils.OTelTraceProvider(pb.CFE_ServiceDesc.ServiceName, "ganache_otelcoll_1:4317")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get OTEL trace provider")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	grpcCfg := &grpcutils.GRPCServerConfig{
		TLSConfig: &grpcutils.TLSConfig{
			CertFilePath:     *tlsCrtPath,
			KeyFilePath:      *tlsKeyPath,
			ClientCAFilePath: *clientCAPath,
			SkipTLS:          *skipTLS,
			RootCAFilePath:   *rootCAPath,
		},
		KeepAliveConfig: &grpcutils.KeepAliveConfig{}, // use defaults
	}

	cfeServer, err := server.New(grpcCfg, *etcdSpec, *csResolverPrefix, *shards)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create CFE server.")
	}
	s, err := grpcCfg.NewGRPCServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load grpc server opts.")
	}

	// register CFE server
	pb.RegisterCFEServer(s, cfeServer)

	log.Info().Msgf("Starting CFE on address %v", lis.Addr().String())
	s.Serve(lis)
}
