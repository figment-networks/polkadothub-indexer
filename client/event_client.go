package client

import (
	"context"
	"google.golang.org/grpc"

	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
)

var (
	_ EventClient = (*eventClient)(nil)
)

type EventClient interface {
	GetByHeight(int64) (*eventpb.GetByHeightResponse, error)
}

func NewEventClient(conn *grpc.ClientConn) *eventClient {
	return &eventClient{
		client: eventpb.NewEventServiceClient(conn),
	}
}

type eventClient struct {
	client eventpb.EventServiceClient
}

func (r *eventClient) GetByHeight(h int64) (*eventpb.GetByHeightResponse, error) {
	ctx := context.Background()

	return r.client.GetByHeight(ctx, &eventpb.GetByHeightRequest{Height: h})
}
