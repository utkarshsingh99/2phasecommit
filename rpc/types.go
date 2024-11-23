package rpcmanager

import (
	"net/rpc"
	"time"

	"github.com/utkarshsingh99/2phasecommit/bank"
)

type RPCManager struct {
	// Map of server names to their ports
	ServerPorts map[string]string
	// Map of client names to their portC
	ClientPorts map[string]string
	// Map of server names to their rpcClients
	RPCClients map[string]*rpc.Client
	// Map of cluster names to server names
	ShardMapping map[string][]string
	// Map of cluster names to contact server
	ShardLeaderMapping map[string]string
}

var ServerClusterMap = map[string]string{
	"S1": "C1",
	"S2": "C1",
	"S3": "C1",
	"S4": "C2",
	"S5": "C2",
	"S6": "C2",
	"S7": "C3",
	"S8": "C3",
	"S9": "C3",
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
	Sender    int
	Receiver  int
	Amount    int
	ID        string
	StartTime time.Time
	LogTime   time.Time
}

type ToggleServerInterface struct {
	From string
	To   string
}

type PrepareInterface struct {
	BallotNumber      int
	LastTransactionId string
}

type SyncInterface struct {
	LastTransactionId string
}

type SyncResponseInterface struct {
	RemainingLog []bank.Transaction
}

type PromiseInterface struct {
	BallotNumber int
	AcceptNumber int
	AcceptValue  Transaction
}

type AcceptInterface struct {
	BallotNumber int
	AcceptNumber int
	AcceptValue  Transaction
}

type AcceptedInterface struct {
	BallotNumber int
	AcceptNumber int
	AcceptValue  Transaction
}

type CommitInterface struct {
	BallotNumber int
	AcceptNumber int
	AcceptValue  Transaction
}
