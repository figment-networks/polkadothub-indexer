package client

import (
	"google.golang.org/grpc"
)

func New(connStr string) (*Client, error) {
	conn, err := grpc.Dial(connStr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,

		Chain:       NewChainClient(conn),
		Account:     NewAccountClient(conn),
		Block:       NewBlockClient(conn),
		Transaction: NewTransactionClient(conn),
	}, nil
}

type Client struct {
	conn *grpc.ClientConn

	Chain       ChainClient
	Account     AccountClient
	Block       BlockClient
	Transaction TransactionClient
}

func (c *Client) Close() error {
	return c.conn.Close()
}
