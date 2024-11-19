package rpcmanager

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

func New() *RPCManager {

	gob.Register(Message{})
	gob.Register(RequestInterface{})
	gob.Register(ReplyInterface{})

	return &RPCManager{
		RPCClients: make(map[string]*rpc.Client),
		ServerPorts: map[string]string{
			"S1": ":1234",
			"S2": ":1235",
			"S3": ":1236",
			"S4": ":1237",
			"S5": ":1238",
			"S6": ":1239",
			"S7": ":1240",
			"S8": ":1241",
			"S9": ":1242",
		},
		ClientPorts: map[string]string{
			"Client": ":2234",
		},
	}
}

func (RPCManager *RPCManager) saveToRPCClientMap(remoteMap map[string]string, currentServerID string) {
	var wg sync.WaitGroup // WaitGroup to wait for all dialing and calling to complete
	mu := &sync.Mutex{}   // Mutex to safely update shared resources like rpcClients

	// Asynchronously dial servers
	for serverID, serverPort := range remoteMap {
		if serverID != currentServerID {
			wg.Add(1) // Increment the WaitGroup counter
			go func(serverID, serverPort string) {
				defer wg.Done() // Decrement the WaitGroup counter when done
				client, err := RPCManager.DialServer(serverID, serverPort)
				if err != nil {
					log.Printf("Error connecting to %s: %v\n", serverID, err)
					return
				}

				// Safely store the rpcClient in the map using a mutex
				mu.Lock()
				RPCManager.RPCClients[serverID] = client
				mu.Unlock()
			}(serverID, serverPort)
		}
	}
	// Wait for all dial operations to complete
	wg.Wait()
	log.Println("Connected to all servers or clients")
}

// Asynchronous function to dial all servers except itself
func (RPCManager *RPCManager) ConnectToServers(currentServerID string) {

	RPCManager.saveToRPCClientMap(RPCManager.ServerPorts, currentServerID)
	RPCManager.saveToRPCClientMap(RPCManager.ClientPorts, currentServerID)

	// Reset WaitGroup for RPC calls
	wg := sync.WaitGroup{}

	// Asynchronously make RPC calls to all connected RPC Clients
	for remoteServerID, remoteServerRPC := range RPCManager.RPCClients {
		if remoteServerRPC != nil {
			wg.Add(1) // Increment the WaitGroup counter
			go func(remoteServerID string, remoteServerRPC *rpc.Client) {
				defer wg.Done() // Decrement the WaitGroup counter when done
				var reply int
				data := Message{
					Type: "Ping",
					From: currentServerID,
				}

				var err error
				// If remoteServerID begins with S, it is server, else it is client
				if remoteServerID[0] == 'S' {
					err = remoteServerRPC.Call("Server.Ping", data, &reply)
				} else {
					err = remoteServerRPC.Call("Client.Ping", data, &reply)
				}
				if err != nil {
					log.Printf("RPC error calling %s: %v\n", remoteServerID, err)
				}
			}(remoteServerID, remoteServerRPC)
		}
	}

	// Wait for all RPC call operations to complete
	wg.Wait()
}

func isServer(serverID string) bool {
	return serverID[0] == 'S'
}

// BroadcastMessage sends an RPC call to all RPCClients asynchronously
// and returns a map of server IDs and their responses.
func (RPCManager *RPCManager) BroadcastMessage(serviceMethod string, message Message) map[string]int {
	replies := make(map[string]int) // Map to store responses from each server
	mu := &sync.Mutex{}             // Mutex to synchronize access to the replies map
	wg := &sync.WaitGroup{}         // WaitGroup to wait for all goroutines to finish

	for serverId, client := range RPCManager.RPCClients {
		if client != nil && isServer(serverId) {
			wg.Add(1)
			// Launch a goroutine to make the RPC call
			go func(serverId string, client *rpc.Client) {
				defer wg.Done()
				var reply int
				err := client.Call(serviceMethod, message, &reply)
				if err != nil {
					log.Printf("RPC error calling %s: %v\n", serverId, err)
					return
				}

				// Safely store the reply using a mutex
				mu.Lock()
				replies[serverId] = reply
				mu.Unlock()

				// log.Println("Response from %s: %d\n", serverId, reply)
			}(serverId, client)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return replies // Return the map with server responses
}

// Dial to another server with retries and store the client in the map
func (RPCManager *RPCManager) DialServer(serverID, serverPort string) (*rpc.Client, error) {
	var client *rpc.Client
	var err error
	maxRetries := 3
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		client, err = rpc.Dial("tcp", "localhost"+serverPort)
		if err == nil {
			// log.Printf("Connected to server %s on port %s!\n", serverID, serverPort)
			return client, nil
		}
		log.Printf("Dialing server %s failed: %v. Retrying in %v seconds...\n", serverID, err, retryDelay.Seconds())
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to server %s after %d retries", serverID, maxRetries)
}

// BroadcastMessage sends an RPC call to all RPCClients asynchronously
// and returns a map of server IDs and their responses.
func (RPCManager *RPCManager) BroadcastSignedMessage(serviceMethod string, message Message) map[string]int {
	replies := make(map[string]int) // Map to store responses from each server
	mu := &sync.Mutex{}             // Mutex to synchronize access to the replies map
	wg := &sync.WaitGroup{}         // WaitGroup to wait for all goroutines to finish

	for serverId, client := range RPCManager.RPCClients {
		if client != nil && isServer(serverId) {
			wg.Add(1)
			// Launch a goroutine to make the RPC call
			go func(serverId string, client *rpc.Client) {
				defer wg.Done()
				var reply int
				err := client.Call(serviceMethod, message, &reply)
				if err != nil {
					log.Printf("RPC error calling %s: %v\n", serverId, err)
					return
				}

				// Safely store the reply using a mutex
				mu.Lock()
				replies[serverId] = reply
				mu.Unlock()

				// log.Println("Response from %s: %d\n", serverId, reply)
			}(serverId, client)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return replies // Return the map with server responses
}
