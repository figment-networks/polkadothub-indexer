package client

import (
	"context"
	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
)

var (
	_ TransactionClient = (*transactionClient)(nil)
)

type TransactionClient interface {
	GetByHeight(int64) (*transactionpb.GetByHeightResponse, error)
}

func NewTransactionClient(conn *grpc.ClientConn) TransactionClient {
	return &transactionClient{
		client: transactionpb.NewTransactionServiceClient(conn),
	}
}

type transactionClient struct {
	client transactionpb.TransactionServiceClient
}

func (r *transactionClient) GetByHeight(h int64) (*transactionpb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &transactionpb.GetByHeightRequest{Height: h})
}

