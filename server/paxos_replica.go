package main

import (
	"log"
	"sync"

	"github.com/utkarshsingh99/2phasecommit/bank"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

// FUNC: PREPARE(<Ballot Number, Last Transaction ID>)
// Received a prepare message from a leader
// If ballot number > AcceptNum
// Compare last transaction ID with the last transaction ID in the log
// If our log is behind, send a GiveSync message to the leader.
// If our log is ahead, send a TakeSync message to the leader.
// Return in either case.

// If log is updated,
// update AcceptNum, reset AcceptVal
// send a Promise message ⟨ACK, Ballot Number, AcceptNum, AcceptVal⟩
// RETURN

func (s *Server) Prepare(message rpcmanager.Message, reply *int) error {
	if !s.Active {
		return nil
	}

	log.Println("Received Prepare from", message.From, "for ballot number", message.Payload.(rpcmanager.PrepareInterface).BallotNumber)

	ballotNumber := message.Payload.(rpcmanager.PrepareInterface).BallotNumber
	lastTransId := message.Payload.(rpcmanager.PrepareInterface).LastTransactionId

	client := s.RPCManager.RPCClients[message.From]

	if s.Paxos.acceptNumber <= ballotNumber {
		prevBallotNumber := s.Paxos.acceptNumber
		s.Paxos.acceptNumber = ballotNumber
		s.Paxos.ballotNumber = ballotNumber

		compareLog, compareLogMsg := s.Bank.Log.CompareMyLog(lastTransId)
		if !compareLog {
			log.Println(compareLog, compareLogMsg)

			syncMessage := rpcmanager.Message{
				From: s.StringId,
				Payload: rpcmanager.SyncInterface{
					LastTransactionId: lastTransId,
				},
			}

			var syncErr error

			if compareLogMsg == "me_outdated" {
				syncErr = client.Call("Server.TakeSync", syncMessage, nil)
			} else if compareLogMsg == "they_outdated" {
				syncErr = client.Call("Server.GiveSync", syncMessage, nil)
			}
			if syncErr != nil {
				log.Println("Error calling Sync:", syncErr)
				return syncErr
			}
		}

		respMessage := rpcmanager.Message{
			From: s.StringId,
			Payload: rpcmanager.PromiseInterface{
				BallotNumber: s.Paxos.ballotNumber,
				AcceptNumber: prevBallotNumber,
				AcceptValue:  s.Paxos.acceptValue,
			},
		}

		log.Println("Sending Promise message...", respMessage)

		// Send Promise Message
		err := client.Call("Server.Promise", respMessage, nil)
		if err != nil {
			log.Println("Error calling Promise:", err)
			return err
		}
	}
	return nil
}

// FUNC: RECEIVE ACCEPT(<Ballot Number, AcceptNum, AcceptVal>)
// Get transaction from AcceptVal
// Acquire lock on data items for transaction
// Send Accepted message <Ballot Number, AcceptNum, AcceptVal=Transaction>
// RETURN
func (s *Server) Accept(message rpcmanager.Message, reply *int) error {
	if !s.Active {
		return nil
	}
	// log.Println("Received Accept from", message.From, message.Payload.(rpcmanager.AcceptInterface))

	transaction := message.Payload.(rpcmanager.AcceptInterface).AcceptValue

	// Acquire locks on x and y
	s.Bank.LockClient(transaction.Sender)
	s.Bank.LockClient(transaction.Receiver)

	respMessage := rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.AcceptedInterface{
			BallotNumber: s.Paxos.ballotNumber,
			AcceptNumber: message.Payload.(rpcmanager.AcceptInterface).AcceptNumber,
			AcceptValue:  transaction,
		},
	}

	// Send Accepted message back to primary
	go s.RPCManager.RPCClients[message.From].Call("Server.Accepted", respMessage, nil)
	return nil
}

// FUNC: COMMIT(<Ballot Number, AcceptNum, AcceptVal>)
// Commit transaction
// Release lock on data items
// RETURN

func (s *Server) Commit(message rpcmanager.Message, reply *int) error {
	if !s.Active {
		return nil
	}
	// log.Println("Received Commit from", message.From, message.Payload.(rpcmanager.CommitInterface))

	transaction := message.Payload.(rpcmanager.CommitInterface).AcceptValue

	s.executeTransaction(transaction)

	s.ResetPaxos()

	// Print balance of x and y
	// log.Println("Balance of x:", s.Bank.GetBalance(transaction.Sender))
	// log.Println("Balance of y:", s.Bank.GetBalance(transaction.Receiver))

	return nil
}

func (s *Server) executeTransaction(transaction rpcmanager.Transaction) {

	// log.Println("Executing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)

	newTransaction := bank.Transaction{
		Sender:    transaction.Sender,
		Receiver:  transaction.Receiver,
		Amount:    transaction.Amount,
		ID:        transaction.ID,
		StartTime: transaction.StartTime,
	}

	s.Bank.CommitTransaction(newTransaction)

	// Release locks on x and y
	s.Bank.UnlockClient(transaction.Sender)
	s.Bank.UnlockClient(transaction.Receiver)
}

func (s *Server) ResetPaxos() {
	s.Paxos.active = false
	s.Paxos.promises = leaderMessages{
		messages: []rpcmanager.Message{},
		lock:     sync.Mutex{},
		acked:    false,
	}
	s.Paxos.accepted = leaderMessages{
		messages: []rpcmanager.Message{},
		lock:     sync.Mutex{},
		acked:    false,
	}
	s.Paxos.acceptValue = rpcmanager.Transaction{}
	s.CurRequest = rpcmanager.Message{}
}
