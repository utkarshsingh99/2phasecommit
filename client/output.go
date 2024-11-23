package main

import (
	"fmt"
	"log"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func displayMenu() int {
	log.Println("Choose an option:")
	log.Println("1. PrintDatastore")
	log.Println("2. PrintBalance")
	log.Println("3. PrintPerformance")
	log.Println("4. Exit")
	fmt.Print("Enter your choice (1-4): ")

	var choice int
	fmt.Scanln(&choice)
	return choice
}

// Function to call PrintBalance, PrintLog, and PrintDB
func (c *Client) callOutputFunctions() {
	for {
		// Display the menu to the user and get their choice
		choice := displayMenu()

		// Validate the user's choice
		switch choice {
		case 1:
			// Print Datastore
			for serverId := range c.RPCManager.RPCClients {
				if serverId[:1] != "S" {
					continue
				}
				client := c.RPCManager.RPCClients[serverId]
				if client == nil {
					log.Println("Client not found for server: ", serverId)
					continue
				}
				log.Println("Calling PrintDatastore for server: ", serverId)
				client.Call("Server.PrintDatastore", rpcmanager.Message{}, nil)
			}
		case 2:
			// Print Balance
			sequenceNum := int(0)
			fmt.Print("Enter data item: ")
			fmt.Scanln(&sequenceNum)
			for serverId := range c.RPCManager.RPCClients {
				if serverId[:1] != "S" {
					continue
				}
				var resp int
				client := c.RPCManager.RPCClients[serverId]
				if client == nil {
					log.Println("Client not found for server: ", serverId)
					continue
				}
				// log.Println("Calling PrintDB for server: ", serverId)
				err := client.Call("Server.PrintBalance", rpcmanager.Message{
					From:    "reader",
					Payload: sequenceNum,
				}, &resp)
				if err != nil {
					log.Println("Error calling PrintDB for server: ", serverId, err)
					continue
				}
				if resp >= 0 {
					fmt.Println("Server", serverId, "has balance: ", resp)
				}
			}
		case 3:
			// Print Performance
			for serverId := range c.RPCManager.RPCClients {
				if serverId[:1] != "S" {
					continue
				}
				client := c.RPCManager.RPCClients[serverId]
				if client == nil {
					log.Println("Client not found for server: ", serverId)
					continue
				}
				// log.Println("Calling PrintPerformance for server: ", serverId)
				reply := []float64{}
				client.Call("Server.PrintPerformance", rpcmanager.Message{}, &reply)
				fmt.Printf("%s -> Average Latency: %.0f µs, Throughput: %.2f /second\n", serverId, reply[1], reply[0])
			}
		case 4:
			// Exit the loop when the user chooses option 5
			log.Println("Exiting...")
			return
		default:
			log.Println("Invalid choice. Please enter a number between 1 and 6.")
		}

		// Reprint the menu after executing a function or an invalid choice
	}
}
