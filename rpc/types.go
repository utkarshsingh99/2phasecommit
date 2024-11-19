package rpcmanager

import (
	"net/rpc"
	"time"
)

type RPCManager struct {
	// Map of server names to their ports
	ServerPorts map[string]string
	// Map of client names to their portC
	ClientPorts map[string]string
	// Map of server names to their rpcClients
	RPCClients map[string]*rpc.Client
}

type Message struct {
	Type    string
	From    string
	Payload interface{}
}

type RequestInterface struct {
	Operation Transaction
	Timestamp time.Time
	Client    string
	RequestID string
}

type ReplyInterface struct {
	Timestamp time.Time
	Client    string
	ReplicaID string
	RequestID string
}

type Transaction struct {
	Sender    string
	Receiver  string
	Amount    int
	ID        string
	StartTime time.Time
	LogTime   time.Time
}
