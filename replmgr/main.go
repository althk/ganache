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

	pb "github.com/althk/ganache/replmgr/proto"
	"github.com/althk/ipmq"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
)

var port = flag.Int("port", 0, "cache server port, defaults to 0 which means any available port")
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

	s := grpc.NewServer()
	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	reflection.Register(s)
	pb.RegisterReplServer(s, &replmgr{
		mq: ipmq.New(),
	})
	log.Info().Msgf("Running repl server on %v", lis.Addr().String())
	s.Serve(lis)
}
