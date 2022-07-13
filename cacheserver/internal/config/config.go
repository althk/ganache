package config

import "github.com/althk/goeasy/grpcutils"

// Cache Server config
type CSConfig struct {
	Port          int32
	CSMSpec       string
	ETCDSpec      string
	MaxCacheBytes int64
	Shard         int32
	Addr          string
	ServerConfig  *grpcutils.GRPCServerConfig
}
