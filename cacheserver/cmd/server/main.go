package main

import (
	"context"
	"flag"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/althk/ganache/cacheserver/internal/config"
	"github.com/althk/ganache/cacheserver/internal/server"
	pb "github.com/althk/ganache/cacheserver/proto"
	grpcutils "github.com/althk/ganache/utils/grpc"
)

var listenAddr = flag.String("listen_addr", ":0", "cache server port, defaults to 0 which means any available port")
var shard = flag.Int("shard", 0, "shard number for key distribution")
var csmSpec = flag.String("csm_server", "", "address of CSM service in the form host:port")
var etcdSpec = flag.String("etcd_server", "localhost:2379", "address of etcd service in the form host:port")
var debug = flag.Bool("debug", false, "enable debug logging")
var maxCacheBytes = flag.Int64("max_cache_bytes", 1000000000, "max size oftotal cache in bytes, defaults to 1GiB")
var clientCAPath = flag.String("client_ca_file", "", "Path to CA cert file that can verify client certs")
var rootCAPath = flag.String("root_ca_file", "", "Path to CA cert file that can verify server/peer certs")
var tlsCrtPath = flag.String("tls_cert_file", "", "Path to server's TLS cert file")
var tlsKeyPath = flag.String("tls_key_file", "", "Path to server's TLS key file")
var skipTLS = flag.Bool("skip_tls", false, "If server should skip TLS and use insecure creds")

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", GetOutboundAddr(*listenAddr))
	if err != nil {
		log.Fatal().Msgf("Error listening on port %v: %v", *listenAddr, err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	shutdownFn, err := grpcutils.OTelTraceProvider(pb.Cache_ServiceDesc.ServiceName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get OTEL trace provider")
	}
	defer func() {
		if err := shutdownFn(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	tlsCfg := &grpcutils.TLSConfig{
		CertFilePath:     *tlsCrtPath,
		KeyFilePath:      *tlsKeyPath,
		ClientCAFilePath: *clientCAPath,
		SkipTLS:          *skipTLS,
		RootCAFilePath:   *rootCAPath,
	}
	csConfig := &config.CSConfig{
		CSMSpec:       *csmSpec,
		ETCDSpec:      *etcdSpec,
		MaxCacheBytes: *maxCacheBytes,
		Shard:         int32(*shard),
		Addr:          lis.Addr().String(),
		TLSConfig:     tlsCfg,
	}
	cacheServer, err := server.New(csConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Cache server initialization failed.")
	}

	// cache server has been registered with CSM and synced the shard locally
	// proceed with serving.
	grpcServerCfg := &grpcutils.GRPCServerConfig{
		TLSConfig: tlsCfg,
	}
	s, err := grpcutils.NewGRPCServer(grpcServerCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load grpc server opts.")
	}
	// register cacheserver
	pb.RegisterCacheServer(s, cacheServer)

	log.Info().Msgf("Running cache server on %v", lis.Addr().String())
	s.Serve(lis)
}

// Get preferred outbound addr (ip:port) of this server
func GetOutboundAddr(addr string) string {
	host, port, _ := net.SplitHostPort(addr)
	if ip := net.ParseIP(host); ip != nil && !ip.IsLoopback() {
		return net.JoinHostPort(ip.String(), port)
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return addr
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return net.JoinHostPort(localAddr.IP.String(), port)
}
