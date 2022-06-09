package main

import (
	"io"

	pb "github.com/althk/ganache/replmgr/proto"
	"github.com/althk/ipmq"
	"github.com/rs/zerolog/log"
)

// replmgr implements ReplServer and syncs
// data between cache servers for HA.
type replmgr struct {
	pb.UnimplementedReplServer
	mq ipmq.MQ
}

func (r *replmgr) Sync(stream pb.Repl_SyncCacheServer) error {

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if in.Register {
			log.Info().
				Str("instance", in.Source).
				Int64("shard", in.Shard).
				Msg("Registered new client")

			c := func(msg ipmq.Msg) error {
				m := msg.(*pb.CacheData)
				err := stream.Send(m)
				return err
			}
			cancel, err := r.mq.Register(c)
			if err != nil {
				return err
			}
			defer cancel()
		}
		r.mq.Push(in)
	}
}
