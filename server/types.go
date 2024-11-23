package main

import (
	"sync"

	bank "github.com/utkarshsingh99/2phasecommit/bank"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

type Server struct {
	StringId     string
	ClusterId    string
	RPCManager   *rpcmanager.RPCManager
	IntId        int
	Active       bool
	Paxos        *Paxos
	Bank         *bank.Bank
	RequestQueue leaderMessages
	CurRequest   rpcmanager.Message
}

type leaderMessages struct {
	messages []rpcmanager.Message
	lock     sync.Mutex
	acked    bool
}

type Paxos struct {
	active       bool
	ballotNumber int
	acceptNumber int
	acceptValue  rpcmanager.Transaction
	promises     leaderMessages
	accepted     leaderMessages
}
