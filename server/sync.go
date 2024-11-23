package main

import (
	"fmt"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

// Create a GiveSync, TakeSync RPC function

// If GiveSync received with a transaction ID argument, send a TakeSync to that server with logs after that ID
func (s *Server) GiveSync(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received GiveSync from", message.From)
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

	err := s.RPCManager.RPCClients[message.From].Call("Server.TakeSyncHandler", respMessage, s.StringId)

	if err != nil {
		fmt.Println("Error calling TakeSyncHandler:", err)
	}

	return nil
}

func (s *Server) TakeSyncHandler(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received TakeSyncHandler from", message.From)
	s.Bank.Log.UpdateLog(message.Payload.(rpcmanager.SyncResponseInterface).RemainingLog)
	// .AppendLog(message.Payload.(rpcmanager.SyncResponseInterface).RemainingLog)
	return nil
}
