package main

import (
	"log"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (c *Client) handleCrossShardTransaction(transaction rpcmanager.Transaction) {
	senderCluster := c.identifyCluster(transaction.Sender)
	receiverCluster := c.identifyCluster(transaction.Receiver)

	senderContact := c.RPCManager.ShardLeaderMapping[senderCluster]
	receiverContact := c.RPCManager.ShardLeaderMapping[receiverCluster]

	RPCFunction := "Server.CrossShardTransaction"

	senderClient := c.RPCManager.RPCClients[senderContact]
	receiverClient := c.RPCManager.RPCClients[receiverContact]

	go func() {
		message := rpcmanager.Message{
			Type: "CrossShardTransaction",
			Payload: rpcmanager.Transaction{
				Sender:    transaction.Sender,
				Receiver:  0,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			},
		}
		err := senderClient.Call(RPCFunction, message, nil)
		if err != nil {
			log.Println("Error calling CrossShardTransaction for sender:", err)
		}
	}()

	go func() {
		message := rpcmanager.Message{
			Type: "CrossShardTransaction",
			Payload: rpcmanager.Transaction{
				Sender:    0,
				Receiver:  transaction.Receiver,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			},
		}
		err := receiverClient.Call(RPCFunction, message, nil)
		if err != nil {
			log.Println("Error calling CrossShardTransaction for receiver:", err)
		}
	}()

}
