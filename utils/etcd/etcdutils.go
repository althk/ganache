package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func V3Client(etcdSpec string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdSpec},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}
