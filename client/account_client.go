package client

import (
	"context"
	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
)

var (
	_ AccountClient = (*accountClient)(nil)
)

type AccountClient interface {
	GetIdentity(string) (*accountpb.GetIdentityResponse, error)
	GetByHeight(string, int64) (*accountpb.GetByHeightResponse, error)
}

func NewAccountClient(conn *grpc.ClientConn) *accountClient {
	return &accountClient{
		client: accountpb.NewAccountServiceClient(conn),
	}
}

type accountClient struct {
	client accountpb.AccountServiceClient
}

func (r *accountClient) GetIdentity(address string) (*accountpb.GetIdentityResponse, error) {
	ctx := context.Background()

	return r.client.GetIdentity(ctx, &accountpb.GetIdentityRequest{Address: address})
}

func (r *accountClient) GetByHeight(address string, h int64) (*accountpb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &accountpb.GetByHeightRequest{
		Address: address,
		Height: h,
	})
}
