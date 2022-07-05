package server

import (
	"time"

	"github.com/althk/ganache/csm/internal/service"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/rs/zerolog/log"
)

func New(etcdSpec, resolverPrefix string) (*service.CSM, error) {
	etcdc, err := etcdV3Client(etcdSpec)
	if err != nil {
		return nil, err
	}
	return service.NewCSM(etcdc, resolverPrefix)
}

func etcdV3Client(etcdSpec string) (*clientv3.Client, error) {
	log.Info().Msgf("Connecting to etcd server: %v", etcdSpec)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil

}
