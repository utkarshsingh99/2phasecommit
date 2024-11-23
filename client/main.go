package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: go run client.go <path-to-csv>")
		return
	}

	if len(os.Args) < 2 {
		log.Println("Usage: go run reader.go <path-to-csv>")
		return
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Start the RPC server in a goroutine
	go startClient("Client", file)

	// Keep the main function running
	select {}
}

func startClient(clientId string, file *os.File) {

	rm := rpcmanager.New()

	client := &Client{
		StringId:   clientId,
		RPCManager: rm,
	}
	fmt.Println("RPC Clients: ", client.RPCManager.RPCClients)
	clientPort := client.RPCManager.ClientPorts[clientId]

	fmt.Println("Starting RPC server on clientPort:", clientPort, client)
	rpc.Register(client)
	fmt.Println("Registered RPC server on clientPort:", clientPort)

	listener, err := net.Listen("tcp", clientPort)
	if err != nil {
		log.Println("Listen error:", err)
		return
	}
	defer listener.Close()

	log.Println("Listening to client requests on clientPort:", clientPort)

	go func() {
		time.Sleep(time.Second * 2)
		log.Println("Connecting to all servers...")
		client.RPCManager.ConnectToServers(clientId)
		log.Println("RPC Client map length: ", len(client.RPCManager.RPCClients))
		if len(client.RPCManager.RPCClients) == 9 {
			log.Println("All servers and clients connected!")
		}
		client.Execute(file)
	}()

	// Logger setup START
	// f, err := os.OpenFile("logs/debug_log_"+clientId+".txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// defer f.Close()

	// log.SetOutput(f)
	// Logger setup END

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
