package client

import (
	"context"

	pb "github.com/althk/ganache/cfe/proto"
	"github.com/althk/goeasy/grpcutils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type CacheClient interface {
	Namespace(ns string)
	SetString(ctx context.Context, k, v string) error
	GetString(ctx context.Context, k string) (string, error)
	SetInt64(ctx context.Context, k string, v int64) error
	GetInt64(ctx context.Context, k string) (int64, error)
	SetMessage(ctx context.Context, k string, msg proto.Message) error
	GetMessage(ctx context.Context, k string, msg proto.Message) error
}

func New(cfeSpec, caFilePath string) (CacheClient, error) {
	tlsCfg := &grpcutils.TLSConfig{
		RootCAFilePath: caFilePath,
		NoClientCert:   true,
	}
	grpcCfg := &grpcutils.GRPCServerConfig{
		TLSConfig:       tlsCfg,
		KeepAliveConfig: &grpcutils.KeepAliveConfig{},
	}
	opts, err := grpcCfg.GetGRPCDialOpts()
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(cfeSpec, opts...)
	if err != nil {
		return nil, err
	}
	cli := pb.NewCFEClient(conn)
	return &client{
		cfe: cli,
	}, err
}

// client implements CacheClient
type client struct {
	ns  string
	cfe pb.CFEClient
}

func (c *client) Namespace(ns string) {
	c.ns = ns
}

func (c *client) set(ctx context.Context, k string, v proto.Message) error {
	d, err := anypb.New(v)
	if err != nil {
		return err
	}
	req := &pb.SetRequest{
		Namespace: c.ns,
		Key:       k,
		Data:      &anypb.Any{},
	}
	proto.Merge(req.Data, d)
	_, err = c.cfe.Set(ctx, req)
	return err
}

func (c *client) get(ctx context.Context, k string, v proto.Message) error {
	req := &pb.GetRequest{
		Namespace: c.ns,
		Key:       k,
	}
	resp, err := c.cfe.Get(ctx, req)
	if err != nil {
		return err
	}
	err = resp.Data.UnmarshalTo(v)
	return err
}

func (c *client) SetString(ctx context.Context, k string, v string) error {
	return c.set(ctx, k, wrapperspb.String(v))
}

func (c *client) GetString(ctx context.Context, k string) (string, error) {
	v := &wrapperspb.StringValue{}
	err := c.get(ctx, k, v)
	return v.Value, err
}

func (c *client) SetInt64(ctx context.Context, k string, v int64) error {
	return c.set(ctx, k, wrapperspb.Int64(v))
}

func (c *client) GetInt64(ctx context.Context, k string) (int64, error) {
	v := &wrapperspb.Int64Value{}
	err := c.get(ctx, k, v)
	return v.Value, err
}

func (c *client) SetMessage(ctx context.Context, k string, msg proto.Message) error {
	return c.set(ctx, k, msg)
}

func (c *client) GetMessage(ctx context.Context, k string, msg proto.Message) error {
	err := c.get(ctx, k, msg)
	return err
}
