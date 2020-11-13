package client

import (
	"google.golang.org/grpc"
)

// maxMsgSize increases the grpc max message size from 4194304 to 419430400
var maxMsgSize = 1024 * 1024 * 400

func New(connStr string) (*Client, error) {
	conn, err := grpc.Dial(
		connStr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,

		Chain:                NewChainClient(conn),
		Account:              NewAccountClient(conn),
		Block:                NewBlockClient(conn),
		Height:               NewHeightClient(conn),
		Transaction:          NewTransactionClient(conn),
		ValidatorPerformance: NewValidatorPerformanceClient(conn),
		Staking:              NewStakingClient(conn),
		Event:                NewEventClient(conn),
		Validator:            NewValidatorClient(conn),
	}, nil
}

type Client struct {
	conn *grpc.ClientConn

	Chain                ChainClient
	Account              AccountClient
	Block                BlockClient
	Height               HeightClient
	Transaction          TransactionClient
	ValidatorPerformance ValidatorPerformanceClient
	Staking              StakingClient
	Event                EventClient
	Validator            ValidatorClient
}

func (c *Client) Close() error {
	return c.conn.Close()
}
