package client

import (
	"context"

	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"google.golang.org/grpc"
)

var (
	_ StakingClient = (*stakingClient)(nil)
)

type StakingClient interface {
	GetByHeight(int64) (*stakingpb.GetByHeightResponse, error)
}

func NewStakingClient(conn *grpc.ClientConn) *stakingClient {
	return &stakingClient{
		client: stakingpb.NewStakingServiceClient(conn),
	}
}

type stakingClient struct {
	client stakingpb.StakingServiceClient
}

func (r *stakingClient) GetByHeight(h int64) (*stakingpb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &stakingpb.GetByHeightRequest{Height: h})
}
