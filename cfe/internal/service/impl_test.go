package service

import (
	"context"
	"testing"

	cspb "github.com/althk/ganache/cacheserver/proto"
	pb "github.com/althk/ganache/cfe/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var setRequestMsg *cspb.SetRequest

func TestNewCFEValid(t *testing.T) {
	cacheClis := make(map[int]cspb.CacheClient)
	cacheClis[0] = &mockCacheClient{}

	got, err := NewCFE(1, cacheClis)
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestNewCFEError(t *testing.T) {
	cacheClis := make(map[int]cspb.CacheClient)
	cacheClis[0] = &mockCacheClient{}

	got, err := NewCFE(0, cacheClis)
	require.Nil(t, got)
	require.ErrorIs(t, err, ErrShardOrClientsEmpty)

	got, err = NewCFE(10, cacheClis)
	require.Nil(t, got)
	require.ErrorIs(t, err, ErrShardAndClientCountsMismatch)
}

func TestCFEGetExistingItem(t *testing.T) {
	cacheClis := make(map[int]cspb.CacheClient)
	cacheClis[0] = &mockCacheClient{}
	c, _ := NewCFE(1, cacheClis)

	resp, err := c.Get(context.TODO(), &pb.GetRequest{
		Namespace: "ns1",
		Key:       "validkey",
	})
	require.NoError(t, err)
	require.EqualValues(t, "someval", resp.Data.Value)
}

func TestCFEGetNonExistentItem(t *testing.T) {
	cacheClis := make(map[int]cspb.CacheClient)
	cacheClis[0] = &mockCacheClient{}
	c, _ := NewCFE(1, cacheClis)

	resp, err := c.Get(context.TODO(), &pb.GetRequest{
		Namespace: "ns1",
		Key:       "nonexistentkey",
	})
	require.Nil(t, resp)
	require.EqualValues(t, codes.NotFound, status.Code(err))
}

func TestCFESet(t *testing.T) {
	cacheClis := make(map[int]cspb.CacheClient)
	cacheClis[0] = &mockCacheClient{}
	c, _ := NewCFE(1, cacheClis)

	_, err := c.Set(context.TODO(), &pb.SetRequest{
		Namespace: "ns1",
		Key:       "validkey",
		Data: &anypb.Any{
			Value: []byte("someval"),
		},
	})
	require.NoError(t, err)
	require.EqualValues(t, []byte("someval"), setRequestMsg.Data.Value)
}

// mockCacheClient implements cspb.CacheClient
type mockCacheClient struct {
	//getReqResponses map[string]
	mock.Mock
}

func (m *mockCacheClient) Get(_ context.Context, in *cspb.GetRequest, _ ...grpc.CallOption) (*cspb.GetResponse, error) {
	k := in.Key
	switch k {
	case "validkey":
		return &cspb.GetResponse{
			Data: &anypb.Any{
				Value: []byte("someval"),
			},
		}, nil
	case "nonexistentkey":
		return nil, status.Error(codes.NotFound, "not found")
	default:
		return nil, status.Error(codes.Unknown, "unknown input")
	}
}
func (m *mockCacheClient) Set(_ context.Context, in *cspb.SetRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	setRequestMsg = &cspb.SetRequest{}
	proto.Merge(setRequestMsg, in)
	return &emptypb.Empty{}, nil
}
func (m *mockCacheClient) Stats(_ context.Context, _ *emptypb.Empty, _ ...grpc.CallOption) (*cspb.StatsResponse, error) {
	return nil, nil
}
