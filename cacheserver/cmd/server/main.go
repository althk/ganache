package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/althk/ganache/cacheserver/config"
	"github.com/althk/ganache/cacheserver/internal/server"
	pb "github.com/althk/ganache/cacheserver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/reflection"

	hpb "google.golang.org/grpc/health/grpc_health_v1"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
var shard = flag.Int("shard", 0, "shard number for key distribution")
var csmSpec = flag.String("csm_server", "", "address of CSM service in the form host:port")
var etcdSpec = flag.String("etcd_server", "localhost:2379", "address of etcd service in the form host:port")
var debug = flag.Bool("debug", false, "enable debug logging")
var maxCacheBytes = flag.Int64("max_cache_bytes", 1000000000, "max size oftotal cache in bytes, defaults to 1GiB")
var clientCAPath = flag.String("client_ca_file", "", "Path to CA cert file that can verify client certs")
var tlsCrtPath = flag.String("tls_cert_file", "", "Path to server's TLS cert file")
var tlsKeyPath = flag.String("tls_key_file", "", "Path to server's TLS key file")

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

	csConfig := &config.CSConfig{
		Port:          int32(*port),
		CSMSpec:       *csmSpec,
		ETCDSpec:      *etcdSpec,
		MaxCacheBytes: *maxCacheBytes,
		Shard:         int32(*shard),
		Addr:          lis.Addr().String(),
	}
	cacheServer, err := server.New(csConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Cache server initialization failed.")
	}

	// cache server has been registered with CSM and synced the shard locally
	// proceed with serving.

	tcreds, err := getGRPCServerCreds(*clientCAPath, *tlsCrtPath, *tlsKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load grpc creds.")
	}
	serverOpts := server.GetGRPCServerOpts()
	serverOpts = append(serverOpts, tcreds)
	s := grpc.NewServer(serverOpts...)
	pb.RegisterCacheServer(s, cacheServer)

	// health and other services.
	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	reflection.Register(s)

	log.Info().Msgf("Running cache server on %v", lis.Addr().String())
	s.Serve(lis)
}

func getGRPCServerCreds(clientCAPath, tlsCrtPath, tlsKeyPath string) (grpc.ServerOption, error) {
	if tlsCrtPath == "" && tlsKeyPath == "" {
		return grpc.Creds(insecure.NewCredentials()), nil
	}
	tlsKeyPair, err := tls.LoadX509KeyPair(tlsCrtPath, tlsKeyPath)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{tlsKeyPair},
		ClientAuth:   tls.NoClientCert,
	}
	if clientCAPath == "" {
		return grpc.Creds(credentials.NewTLS(cfg)), nil
	}
	clientCA, err := ioutil.ReadFile(clientCAPath)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(clientCA) {
		return nil, fmt.Errorf("error adding client CA cert to pool")
	}
	cfg.ClientAuth = tls.RequireAndVerifyClientCert
	cfg.ClientCAs = certPool

	return grpc.Creds(credentials.NewTLS(cfg)), nil
}
