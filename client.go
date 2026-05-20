package main

import (
	"context"
	"net"
)

type Username string
type Client struct {
	name Username
	conn net.Conn
}

func NewClient(name Username, conn net.Conn) *Client {
	return &Client{name, conn}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		c.conn.Close()
		return
	}
}
