package grpcclient

import (
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
}

type Config struct {
	Url string
}

func New(config Config) *Client {
	cc, err := grpc.Dial(config.Url, grpc.WithInsecure())
	if err != nil {
		panic("could not connect to gRPC proxy")
	}

	return &Client{
		conn: cc,
	}
}

func (c *Client) Client() *grpc.ClientConn {
	return c.conn
}
