package main

import (
	"fmt"
	"time"

	"github.com/utkarshsingh99/2phasecommit/bank"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

func (s *Server) PrintBalance(message rpcmanager.Message, reply *int) error {

	clientId := message.Payload.(int)
	balance := s.Bank.GetBalance(clientId)

	*reply = balance
	return nil
}

func (s *Server) PrintDatastore(message rpcmanager.Message, reply *[]bank.Transaction) error {
	// datastore := s.Bank.Log.Transactions

	datastore := s.Bank.DataStore.Entries
	fmt.Println("Datastore of ", s.StringId, ": ")

	for _, transaction := range datastore {
		tr := rpcmanager.Transaction{
			Sender:   transaction.Sender,
			Receiver: transaction.Receiver,
			Amount:   transaction.Amount,
		}
		fmt.Print(transaction.BallotNumber)
		if transaction.CrossShardStatus != "INTRA" {
			fmt.Print(" ", transaction.CrossShardStatus)
		}
		fmt.Print(" ", displayTransaction(tr), "\n")
	}
	*reply = nil
	return nil
}

func (s *Server) PrintPerformance(message rpcmanager.Message, reply *[]float64) error {

	datastore := s.Bank.Log.Transactions

	var totalLatency time.Duration
	for i := 0; i < len(datastore); i++ {
		latency := datastore[i].LogTime.Sub(datastore[i].StartTime)
		totalLatency += latency
	}
	throughput := 1.0 / float64(totalLatency.Seconds())
	// fmt.Println("Total latency: ", totalLatency.Microseconds(), throughput)
	*reply = []float64{throughput, float64(totalLatency.Microseconds()) / float64(len(datastore))}

	return nil
}
