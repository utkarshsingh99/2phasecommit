package main

import (
	"fmt"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

// Create a GiveSync, TakeSync RPC function

// If GiveSync received with a transaction ID argument, send a TakeSync to that server with logs after that ID
func (s *Server) GiveSync(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received GiveSync from", message.From)

	// Get my last transaction ID

	lastTransId, _ := s.Bank.Log.GetLastTransactionID()

	// Send a TakeSync to that server with logs after that ID

	respMessage := rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.SyncInterface{
			LastTransactionId: lastTransId,
		},
	}
	client := s.RPCManager.RPCClients[message.From]

	err := client.Call("Server.TakeSync", respMessage, nil)

	if err != nil {
		fmt.Println("Error calling TakeSync:", err)
	}

	return nil
}

// If TakeSync received with transaction array, append all transactions to the log
func (s *Server) TakeSync(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received TakeSync from", message.From)

	lastTransId := message.Payload.(rpcmanager.SyncInterface).LastTransactionId

	remainingTransactions := s.Bank.Log.SendRemainingLog(lastTransId)

	respMessage := rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.SyncResponseInterface{
			RemainingLog: remainingTransactions,
		},
	}
	fmt.Println("Sending TakeSyncHandler to", remainingTransactions)
	err := s.RPCManager.RPCClients[message.From].Call("Server.TakeSyncHandler", respMessage, nil)

	if err != nil {
		fmt.Println("Error calling TakeSyncHandler:", err)
	}

	return nil
}

func (s *Server) TakeSyncHandler(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received TakeSyncHandler from", message.From)
	// s.Bank.Log.UpdateLog(message.Payload.(rpcmanager.SyncResponseInterface).RemainingLog)

	for _, transaction := range message.Payload.(rpcmanager.SyncResponseInterface).RemainingLog {
		// Acquire locks on x and y
		fmt.Println("Committing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)
		s.Bank.LockClient(transaction.Sender)
		s.Bank.LockClient(transaction.Receiver)
		s.Bank.CommitTransaction(transaction)
	}
	// .AppendLog(message.Payload.(rpcmanager.SyncResponseInterface).RemainingLog)
	return nil
}
