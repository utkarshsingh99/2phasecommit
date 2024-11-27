package main

import (
	"sync"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

type Client struct {
	StringId       string
	RPCManager     *rpcmanager.RPCManager
	CSTMapping     CSTMap
	TransactionMap TransactionMap
}

type CSTMap struct {
	mapping map[string][]bool
	lock    sync.Mutex
}

type TransactionMap struct {
	mapping map[string]rpcmanager.Transaction
	lock    sync.Mutex
}
