package main

import (
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (c *Client) Ping(data rpcmanager.Message, reply *int) error {
	// log.Println("Received ping from", data.From)
	return nil
}
