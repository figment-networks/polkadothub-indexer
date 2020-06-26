package client

import (
	"context"
	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
)

var (
	_ BlockClient = (*blockClient)(nil)
)

type BlockClient interface {
	GetByHeight(int64) (*blockpb.GetByHeightResponse, error)
}

func NewBlockClient(conn *grpc.ClientConn) *blockClient {
	return &blockClient{
		client: blockpb.NewBlockServiceClient(conn),
	}
}

type blockClient struct {
	client blockpb.BlockServiceClient
}

func (r *blockClient) GetByHeight(h int64) (*blockpb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &blockpb.GetByHeightRequest{Height: h})
}
