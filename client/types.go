package main

import rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"

type Client struct {
	StringId   string
	RPCManager *rpcmanager.RPCManager
}
