package client

import (
	"context"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"google.golang.org/grpc"

)

var (
	_ ValidatorPerformanceClient = (*validatorPerformanceClient)(nil)
)

type ValidatorPerformanceClient interface {
	GetByHeight(int64) (*validatorperformancepb.GetByHeightResponse, error)
}

func NewValidatorPerformanceClient(conn *grpc.ClientConn) *validatorPerformanceClient {
	return &validatorPerformanceClient{
		client: validatorperformancepb.NewValidatorPerformanceServiceClient(conn),
	}
}

type validatorPerformanceClient struct {
	client validatorperformancepb.ValidatorPerformanceServiceClient
}

func (r *validatorPerformanceClient) GetByHeight(h int64) (*validatorperformancepb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &validatorperformancepb.GetByHeightRequest{Height: h})
}
