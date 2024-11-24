package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/utkarshsingh99/2phasecommit/bank"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: go run server.go <server-id>")
		return
	}

	serverId := os.Args[1]

	// Logger setup START
	f, err := os.OpenFile("logs/debug_log_"+serverId+".txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	// Logger setup END

	// Start the RPC server in a goroutine
	go startServer(serverId)

	// Keep the main function running
	select {}
}

func startServer(serverId string) {
	serverIntId, _ := strconv.Atoi(serverId[1:])

	clusterId := rpcmanager.ServerClusterMap[serverId]

	dataItemLimits := bank.ClusterDataMap[clusterId]
	serverBank := bank.New(dataItemLimits[0], dataItemLimits[1])

	server := &Server{
		IntId:      serverIntId,
		StringId:   serverId,
		ClusterId:  clusterId,
		Active:     true,
		RPCManager: rpcmanager.New(),
		Bank:       serverBank,
		Paxos: &Paxos{
			active:       false,
			ballotNumber: 0,
			acceptNumber: 0,
			acceptValue:  rpcmanager.Transaction{},
			promises: leaderMessages{
				messages: []rpcmanager.Message{},
				lock:     sync.Mutex{},
				acked:    false,
			},
			accepted: leaderMessages{
				messages: []rpcmanager.Message{},
				lock:     sync.Mutex{},
				acked:    false,
			},
		},
		RequestQueue: leaderMessages{
			messages: []rpcmanager.Message{},
			lock:     sync.Mutex{},
		},
	}

	port := server.RPCManager.ServerPorts[serverId]

	rpc.Register(server)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Println("Listen error:", err)
		return
	}
	defer listener.Close()

	log.Println("Listening to server on port:", port)

	go func() {
		time.Sleep(time.Second * 2)
		// log.Println("Connecting to all servers...")
		server.RPCManager.ConnectToServers(serverId)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
