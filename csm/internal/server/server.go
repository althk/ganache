package server

import (
	"github.com/althk/ganache/csm/internal/service"
	etcdutils "github.com/althk/ganache/utils/etcd"
)

func New(etcdSpec, resolverPrefix string) (*service.CSM, error) {
	etcdc, err := etcdutils.V3Client(etcdSpec)
	if err != nil {
		return nil, err
	}
	return service.NewCSM(etcdc, resolverPrefix)
}
