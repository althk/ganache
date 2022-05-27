package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

var port = flag.Int("port", 40001, "cache server port")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Error listening on port %v: %v", *port, err)
	}

	s := grpc.NewServer()
	registerCacheServer(s, newCacheServer())
	s.Serve(lis)
}
