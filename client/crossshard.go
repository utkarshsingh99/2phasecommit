package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (c *Client) getRemoteServers(transaction rpcmanager.Transaction) (*rpc.Client, *rpc.Client) {
	senderCluster := c.identifyCluster(transaction.Sender)
	receiverCluster := c.identifyCluster(transaction.Receiver)

	senderContact := c.RPCManager.ShardLeaderMapping[senderCluster]
	receiverContact := c.RPCManager.ShardLeaderMapping[receiverCluster]

	senderClient := c.RPCManager.RPCClients[senderContact]
	receiverClient := c.RPCManager.RPCClients[receiverContact]

	return senderClient, receiverClient
}

func (c *Client) sendCSMessage(transaction rpcmanager.Transaction, client *rpc.Client, clientProp string) {
	RPCFunction := "Server.CrossShardTransaction"

	var message rpcmanager.Message
	if clientProp == "sender" {
		message = rpcmanager.Message{
			Type: "CrossShardTransaction",
			Payload: rpcmanager.Transaction{
				Sender:    transaction.Sender,
				Receiver:  0,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			},
		}
	} else if clientProp == "receiver" {
		message = rpcmanager.Message{
			Type: "CrossShardTransaction",
			Payload: rpcmanager.Transaction{
				Sender:    0,
				Receiver:  transaction.Receiver,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			},
		}
	}
	err := client.Call(RPCFunction, message, nil)
	if err != nil {
		log.Println("Error calling CrossShardTransaction for ", clientProp, ":", err)
	}
}

func (c *Client) handleCrossShardTransaction(transaction rpcmanager.Transaction) {

	senderClient, receiverClient := c.getRemoteServers(transaction)

	// Add transaction to CSTMapping
	c.CSTMapping.lock.Lock()
	c.CSTMapping.mapping[transaction.ID] = []bool{}
	c.CSTMapping.lock.Unlock()

	// Add transaction to TransactionMap
	c.TransactionMap.lock.Lock()
	c.TransactionMap.mapping[transaction.ID] = transaction
	c.TransactionMap.lock.Unlock()

	go c.sendCSMessage(transaction, senderClient, "sender")

	go c.sendCSMessage(transaction, receiverClient, "receiver")
}

func (c *Client) sendFinalCSMessage(message rpcmanager.Message, client *rpc.Client, clientProp string) {
	RPCFunction := "Server.FinalCrossShardTransactionStatus"

	err := client.Call(RPCFunction, message, nil)
	if err != nil {
		log.Println("Error calling FinalCrossShardTransactionStatus for ", clientProp, ":", err)
	}
}

func (c *Client) TransactionStatus(data rpcmanager.Message, reply *int) error {

	status := data.Payload.(rpcmanager.StatusInterface).Status
	transaction := data.Payload.(rpcmanager.StatusInterface).Transaction

	fmt.Println("Received transaction status from", data.From)
	c.CSTMapping.lock.Lock()
	defer c.CSTMapping.lock.Unlock()
	if status == "PREPARED" {
		c.CSTMapping.mapping[transaction.ID] = append(c.CSTMapping.mapping[transaction.ID], true)
	} else if status == "ABORT" {
		c.CSTMapping.mapping[transaction.ID] = append(c.CSTMapping.mapping[transaction.ID], false)
	}

	// fmt.Println("CSTMapping length: ", len(c.CSTMapping.mapping[transaction.ID]))
	if len(c.CSTMapping.mapping[transaction.ID]) == 2 {
		// Send commit/abort message to both server and client
		statusMessage := "COMMIT"
		for _, status := range c.CSTMapping.mapping[transaction.ID] {
			if !status {
				statusMessage = "ABORT"
			}
		}

		actualTransaction := c.TransactionMap.mapping[transaction.ID]

		message := rpcmanager.Message{
			From: c.StringId,
			Payload: rpcmanager.StatusInterface{
				Transaction: actualTransaction,
				Status:      statusMessage,
			},
		}
		fmt.Println("Sending final commit/abort message to both sender and receiver", statusMessage)
		senderClient, receiverClient := c.getRemoteServers(actualTransaction)

		time.Sleep(time.Second * 2)

		go c.sendFinalCSMessage(message, senderClient, "sender")
		go c.sendFinalCSMessage(message, receiverClient, "receiver")

	}
	return nil
}
