package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (c *Client) Execute(file *os.File) {

	// // Dial other servers and store clients in the map

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	currentSetNumber := 1
	transactionSet := []rpcmanager.Transaction{}
	activeServersStr := ""
	byzantineServersStr := ""

	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file reached
		}

		setNumberStr := strings.TrimSpace(record[0])
		if setNumberStr != "" {
			// New set detected
			setNumber, _ := strconv.Atoi(setNumberStr)

			// If we detect a new set number, we must process the previous set first
			if setNumber != currentSetNumber {
				// Process the current set of transactions
				log.Println("Transaction set: ", transactionSet)

				c.processTransactionSet(transactionSet, activeServersStr, byzantineServersStr)

				// Prompt user to continue to the next set
				log.Println("Press 'Enter' to continue to the next set of transactions...")
				fmt.Scanln()

				// Reset the transaction set and update the current set number
				transactionSet = []rpcmanager.Transaction{}
				currentSetNumber = setNumber
				activeServersStr = strings.Trim(record[2], "[]")
				byzantineServersStr = strings.Trim(record[3], "[]")

			}
		}

		// Parse transaction details
		transactionDetails := strings.Trim(record[1], "()")
		transactionParts := strings.Split(transactionDetails, ", ")

		amount, _ := strconv.Atoi(transactionParts[2])

		senderInt, _ := strconv.Atoi(transactionParts[0])
		receiverInt, _ := strconv.Atoi(transactionParts[1])

		transaction := rpcmanager.Transaction{
			Sender:   senderInt,
			Receiver: receiverInt,
			Amount:   amount,
		}

		// Add the transaction to the current set
		transactionSet = append(transactionSet, transaction)

		if activeServersStr == "" && record[2] != "" {
			activeServersStr = strings.Trim(record[2], "[]")
		}
		if byzantineServersStr == "" && record[3] != "" {
			byzantineServersStr = strings.Trim(record[3], "[]")
		}
	}

	// After the loop, process the final set of transactions
	if len(transactionSet) > 0 {
		c.processTransactionSet(transactionSet, activeServersStr, byzantineServersStr)
	}

	for serverId := range c.RPCManager.RPCClients {
		// Call 3 functions for each server
		client := c.RPCManager.RPCClients[serverId]
		if client == nil {
			log.Println("Client not found for server: ", serverId)
			continue
		}
		log.Println("Calling output functions for server: ", serverId)
		client.Call("Server.PrintBalance", rpcmanager.Message{}, nil)
		client.Call("Server.PrintLog", rpcmanager.Message{}, nil)
		client.Call("Server.PrintDB", rpcmanager.Message{}, nil)
	}
}

func (c *Client) toggleServerLiveness(servers []string, enable bool) {

	var rpcFunc string
	if enable {
		rpcFunc = "Server.EnableServer"
		log.Println("Enabling servers: ", servers)
	} else {
		rpcFunc = "Server.DisableServer"
		log.Println("Disabling servers: ", servers)
	}

	for _, server := range servers {
		client := c.RPCManager.RPCClients[server]
		if client == nil {
			continue
		}
		resp := 0
		err := client.Call(rpcFunc, rpcmanager.Message{
			Type: "ToggleServer",
			Payload: rpcmanager.ToggleServerInterface{
				From: "reader",
				To:   server,
			},
		}, &resp)
		if err != nil {
			log.Println("Error calling EnableServer for server:", server, err)
			return
		}
	}
}

func (c *Client) updateShardMapping(servers []string) {
	clusters := []string{"C1", "C2", "C3"}
	for idx, server := range servers {
		c.RPCManager.UpdateShardMapping(clusters[idx], server)
	}
}

func (c *Client) identifyCluster(dataItem int) string {

	if dataItem < 1000 {
		return "C1"
	} else if dataItem < 2000 {
		return "C2"
	} else {
		return "C3"
	}
}

func (c *Client) processTransactionSet(transactionSet []rpcmanager.Transaction, activeServersStr string, contactServerStr string) {

	allServers := []string{"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8", "S9"}
	serversToDisable := []string{}

	// Disable all servers not in activeServersStr
	for _, serverId := range allServers {
		if !strings.Contains(activeServersStr, serverId) {
			serversToDisable = append(serversToDisable, serverId)
		}
	}
	c.toggleServerLiveness(serversToDisable, false)

	// Enable all servers in activeServersStr
	c.toggleServerLiveness(strings.Split(activeServersStr, ", "), true)

	var contactServers []string
	if contactServerStr != "" {
		contactServers = strings.Split(contactServerStr, ", ")
	} else {
		contactServers = []string{}
	}
	// Enable byzantine servers
	fmt.Println("Contact servers:", contactServers)
	c.updateShardMapping(contactServers)
	time.Sleep(time.Second * 2)

	// Create a channel for each sender and a goroutine to process that channel sequentially
	for _, transaction := range transactionSet {
		fmt.Println("Processing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)

		transactionObj := rpcmanager.Transaction{
			Sender:    transaction.Sender,
			Receiver:  transaction.Receiver,
			Amount:    transaction.Amount,
			ID:        uuid.New().String(),
			StartTime: time.Now(),
		}

		senderCluster := c.identifyCluster(transaction.Sender)
		receiverCluster := c.identifyCluster(transaction.Receiver)

		if senderCluster != receiverCluster {
			fmt.Println("Cross shard transaction! ", senderCluster, receiverCluster)

			go c.handleCrossShardTransaction(transactionObj)
		} else {
			fmt.Println("Intra shard transaction! ", senderCluster, receiverCluster)
			message := rpcmanager.Message{
				Type:    "IntraShardTransaction",
				From:    "Client",
				Payload: transactionObj,
			}

			go func() {
				contactServer := c.RPCManager.ShardLeaderMapping[senderCluster]

				client := c.RPCManager.RPCClients[contactServer]
				if client == nil {
					log.Println("Client not found for server: ", contactServer)
					return
				}
				err := client.Call("Server.IntraShardTransaction", message, nil)
				if err != nil {
					fmt.Println("Error calling IntraShardTransaction for server:", contactServer, err)
					return
				}
			}()
		}
	}
	// 	clientName := transaction.Sender
	// 	if _, ok := transactionChannels[clientName]; !ok {
	// 		// Create a buffered channel for each sender
	// 		transactionChannels[clientName] = make(chan rpcmanager.Transaction, len(transactionSet))

	// 		// Start a goroutine for each sender to process transactions sequentially
	// 		go func(clientName string, ch chan rpcmanager.Transaction) {
	// 			log.Println("Starting goroutine for client: ", clientName, len(ch))
	// 			clientId := letterMap[clientName]

	// 			for transaction := range ch {
	// 				// Make RPC call to client
	// 				client := RPCManager.RPCClients[clientId]
	// 				log.Println("Calling Invoke for server: ", clientId, " for transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)
	// 				if client == nil {
	// 					log.Println("Client not found for server: ", clientId)
	// 					continue
	// 				}

	// 				message := rpcmanager.Message{
	// 					Type: "Invoke",
	// 					Payload: rpcmanager.InvokeInterface{
	// 						Transaction: transaction,
	// 						Client:      clientId,
	// 					},
	// 				}

	// 				err := client.Call("Client.Invoke", message, nil)
	// 				if err != nil {
	// 					log.Println("Error calling Invoke for server:", clientId, err)
	// 					continue
	// 				}

	// 				// invoke(transaction, clientId) // Process each transaction sequentially for the sender
	// 			}
	// 		}(clientName, transactionChannels[clientName])
	// 	}
	// }

	// No need to wait for goroutines to finish, the main thread can continue immediately
	// Prompt user to continue to the next set
	log.Println("Press 'Enter' to continue call outputs...")
	fmt.Scanln()

	c.callOutputFunctions()
}
