package main

import (
	"fmt"
	"log"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (a *Server) EnableServer(msg *rpcmanager.Message, reply *int) error {
	log.Println("Enabling server: ", a.StringId)
	a.Active = true
	*reply = 1
	return nil
}

func (a *Server) DisableServer(msg *rpcmanager.Message, reply *int) error {
	log.Println("Disabling server: ", a.StringId)
	*reply = 1
	a.Active = false
	return nil
}

func (s *Server) Ping(data rpcmanager.Message, reply *int) error {
	// fmt.Println("Received ping from", data.From)
	return nil
}

func displayTransaction(transaction rpcmanager.Transaction) string {
	return fmt.Sprintf("{ %d, %d, %d }", transaction.Sender, transaction.Receiver, transaction.Amount)
}
