package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
)

var port = flag.Int("port", 40001, "cache server port")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Error listening on port %v: %v", *port, err)
	}

	s := grpc.NewServer()
	h := health.NewServer()
	hpb.RegisterHealthServer(s, h)
	registerCacheServer(s, newCacheServer(h))
	log.Printf("Running cache server on %v\n", lis.Addr().String())
	s.Serve(lis)
}
