package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/height/heightpb"
)

var (
	_ HeightClient = (*heightClient)(nil)
)

type HeightClient interface {
	GetAll(int64) (*heightpb.GetAllResponse, error)
}

func NewHeightClient(conn *grpc.ClientConn) *heightClient {
	return &heightClient{
		client: heightpb.NewHeightServiceClient(conn),
	}
}

type heightClient struct {
	client heightpb.HeightServiceClient
}

func (r *heightClient) GetAll(h int64) (*heightpb.GetAllResponse, error) {
	ctx := context.Background()

	return r.client.GetAll(ctx, &heightpb.GetAllRequest{Height: h})
}
