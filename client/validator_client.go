package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/validator/validatorpb"
)

var (
	_ ValidatorClient = (*validatorClient)(nil)
)

type ValidatorClient interface {
	GetByHeight(int64) (*validatorpb.GetAllByHeightResponse, error)
}

func NewValidatorClient(conn *grpc.ClientConn) *validatorClient {
	return &validatorClient{
		client: validatorpb.NewValidatorServiceClient(conn),
	}
}

type validatorClient struct {
	client validatorpb.ValidatorServiceClient
}

func (r *validatorClient) GetByHeight(h int64) (*validatorpb.GetAllByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetAllByHeight(ctx, &validatorpb.GetAllByHeightRequest{Height: h})
}
