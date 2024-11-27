package main

import (
	"log"
	"math"
	"time"

	"github.com/utkarshsingh99/2phasecommit/bank"
	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

// Contact Server
var noMajorityChecker rpcmanager.Transaction

func (s *Server) IntraShardTransaction(message rpcmanager.Message, reply *int) error {
	log.Println("Received IntraShardTransaction from", message.From, displayTransaction(message.Payload.(rpcmanager.Transaction)))

	// Check if message exists in requestQueue already
	for _, msg := range s.RequestQueue.messages {
		if msg.Payload.(rpcmanager.Transaction).ID == message.Payload.(rpcmanager.Transaction).ID {
			return nil
		}
	}

	s.RequestQueue.lock.Lock()

	// Append message to request Queue
	s.RequestQueue.messages = append(s.RequestQueue.messages, message)

	s.RequestQueue.lock.Unlock()

	// Initiate Request
	s.InitiateRequest()
	return nil
}

func (s *Server) CrossShardTransaction(message rpcmanager.Message, reply *int) error {
	log.Println("Received CrossShardTransaction from", message.From, displayTransaction(message.Payload.(rpcmanager.Transaction)))

	// Check if message exists in requestQueue already
	for _, msg := range s.RequestQueue.messages {
		if msg.Payload.(rpcmanager.Transaction).ID == message.Payload.(rpcmanager.Transaction).ID {
			return nil
		}
	}

	s.RequestQueue.lock.Lock()

	// Append message to request Queue
	s.RequestQueue.messages = append(s.RequestQueue.messages, message)

	s.RequestQueue.lock.Unlock()

	// Initiate Request
	s.InitiateRequest()
	return nil
}

// FUNC: Initiate REQUEST(<Transaction>)
// Contact server receives a transaction request
// Increases ballot number
// Send prepare message to all servers in cluster <Ballot Number, last Transaction ID>
// FUNC: SEND PREPARE(<Ballot Number, Last Transaction ID>)
// Send a prepare message to a server
// RETURN

func (s *Server) InitiateRequest() {
	s.RequestQueue.lock.Lock()
	defer s.RequestQueue.lock.Unlock()
	log.Println("Initiate Request check: ", s.Paxos.active, s.CurRequest.Payload, len(s.RequestQueue.messages))
	if s.Paxos.active || s.CurRequest.Payload != nil || len(s.RequestQueue.messages) == 0 {
		return // Already in a request
	}

	log.Println("Initiating request for...", s.RequestQueue.messages)
	// Pop first element from request queue
	s.CurRequest = s.RequestQueue.messages[0]
	s.RequestQueue.messages = s.RequestQueue.messages[1:]

	if s.CurRequest.Payload == nil {
		return
	}

	// Increase ballot number
	// a.ballotNumber = math.Floor(a.ballotNumber) + 1 + 0.1*float64(a.intId)
	s.Paxos.ballotNumber = math.Floor(s.Paxos.ballotNumber) + 1 + 0.1*float64(s.IntId)
	// s.Paxos.ballotNumber++
	s.Paxos.active = true

	time.Sleep(time.Second)
	lastTransId, _ := s.Bank.Log.GetLastTransactionID()

	// Broadcast prepare message to all servers in cluster
	go s.RPCManager.BroadcastMessage("Server.Prepare", rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.PrepareInterface{
			BallotNumber:      s.Paxos.ballotNumber,
			LastTransactionId: lastTransId,
		},
	}, s.ClusterId)

	// Add timer to abort request in case of no majority
	noMajorityChecker = s.CurRequest.Payload.(rpcmanager.Transaction)
	time.AfterFunc(time.Second*5, func() {
		log.Println("Checking for timeout: ", noMajorityChecker, s.CurRequest.Payload)
		if s.CurRequest.Payload != nil && noMajorityChecker == s.CurRequest.Payload.(rpcmanager.Transaction) {
			log.Println("Aborting request due to timeout")
			s.CurRequest = rpcmanager.Message{}
			s.ResetPaxos()
		}
	})
}

// FUNC: RECEIVE PROMISE(<ACK, Ballot Number, AcceptNum, AcceptVal>)
// Received a promise message from a server
// If number of promise messages received > N/2
// CHECK: (1) there are no locks on data items x and y
// CHECK: (2) the balance of x is at least equal to amt.
// ACQUIRE LOCKS on x and y
// Send Accept message <Ballot Number, AcceptNum, AcceptVal=Transaction>
// RETURN

func (s *Server) Promise(message rpcmanager.Message, reply *int) error {

	// log.Println("Ballot numbers: ", s.Paxos.ballotNumber, message.Payload.(rpcmanager.PromiseInterface).BallotNumber)
	if message.Payload.(rpcmanager.PromiseInterface).BallotNumber < s.Paxos.ballotNumber {
		return nil
	}

	s.Paxos.promises.lock.Lock()
	log.Println("Received Promise from", message.From, s.Paxos.promises.acked)
	s.Paxos.promises.messages = append(s.Paxos.promises.messages, message)

	// >= 1 signifies one other server responded
	if len(s.Paxos.promises.messages) >= 1 && !s.Paxos.promises.acked && s.CurRequest.Payload != nil {
		s.Paxos.promises.acked = true

		transaction := s.CurRequest.Payload.(rpcmanager.Transaction)

		sender := transaction.Sender
		receiver := transaction.Receiver

		// Check if there are no locks on data items x and y
		if s.Bank.IsClientLocked(sender) || s.Bank.IsClientLocked(receiver) {
			// Abort transaction
			s.CurRequest = rpcmanager.Message{}
			s.ResetPaxos()

			go s.InitiateRequest()
			log.Println("Aborting due to locks unavailability")
			s.sendAbort(transaction)
			return nil
		}

		// Check if the balance of x is at least equal to amt.
		if s.Bank.GetBalance(sender) != -1 && s.Bank.GetBalance(sender) < transaction.Amount {
			// Abort transaction
			s.CurRequest = rpcmanager.Message{}
			s.ResetPaxos()

			go s.InitiateRequest()
			log.Println("Aborting due to insufficient balance")
			s.sendAbort(transaction)
			return nil
		}

		// Acquire locks on x and y
		s.Bank.LockClient(sender)
		s.Bank.LockClient(receiver)

		s.Paxos.acceptNumber = s.Paxos.ballotNumber
		s.Paxos.acceptValue = transaction

		log.Println("Broadcasting Accept")
		// Send Accept message
		go s.RPCManager.BroadcastMessage("Server.Accept", rpcmanager.Message{
			From: s.StringId,
			Payload: rpcmanager.AcceptInterface{
				BallotNumber: s.Paxos.ballotNumber,
				AcceptNumber: s.Paxos.acceptNumber,
				AcceptValue:  transaction,
			},
		}, s.ClusterId)

	}
	s.Paxos.promises.lock.Unlock()

	return nil
}

// FUNC: RECEIVE ACCEPTED(<Ballot Number, AcceptNum, AcceptVal=Transaction>)
// If number of accepted messages received > N/2
// Send Commit message <Ballot Number, AcceptNum, AcceptVal=Transaction>
// CALL: COMMIT(<Ballot Number, AcceptNum, AcceptVal=Transaction>) for self
// RETURN

func (s *Server) Accepted(message rpcmanager.Message, reply *int) error {

	log.Println("Received Accepted from", message.From)
	s.Paxos.accepted.lock.Lock()
	s.Paxos.accepted.messages = append(s.Paxos.accepted.messages, message)

	if len(s.Paxos.accepted.messages) >= 1 && s.Paxos.acceptValue.Amount > 0 && !s.Paxos.accepted.acked {

		s.Paxos.accepted.acked = true

		// time.Sleep(time.Second)
		// Send Commit message
		go s.RPCManager.BroadcastMessage("Server.Commit", rpcmanager.Message{
			From: s.StringId,
			Payload: rpcmanager.CommitInterface{
				BallotNumber: s.Paxos.ballotNumber,
				AcceptNumber: s.Paxos.acceptNumber,
				AcceptValue:  s.Paxos.acceptValue,
			},
		}, s.ClusterId)

		s.executeTransaction(s.Paxos.acceptValue)

		s.Paxos.accepted.lock.Unlock()

		transaction := s.Paxos.acceptValue
		s.ResetPaxos()

		bankTrans := bank.Transaction{
			Sender:    transaction.Sender,
			Receiver:  transaction.Receiver,
			Amount:    transaction.Amount,
			ID:        transaction.ID,
			StartTime: transaction.StartTime,
		}

		if transaction.Sender == 0 || transaction.Receiver == 0 {

			// Write in write ahead log
			bankTransaction := bank.Transaction{
				Sender:    transaction.Sender,
				Receiver:  transaction.Receiver,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			}
			s.Bank.WAL.Append(bankTransaction)

			s.Bank.DataStore.AddTransaction(bankTrans, s.Paxos.acceptNumber, "PREPARED")

			go s.InitiateRequest()
			// Send transaction status to client
			go func() {
				message := rpcmanager.Message{
					From: s.StringId,
					Payload: rpcmanager.StatusInterface{
						Transaction: transaction,
						Status:      "PREPARED",
					},
				}
				log.Println("Sending PREPARED message to client")
				err := s.RPCManager.RPCClients["Client"].Call("Client.TransactionStatus", message, nil)
				if err != nil {
					log.Println("Error sending PREPARED message to client:", err)
				}
			}()
		} else {
			// Release locks on x and y
			s.Bank.DataStore.AddTransaction(bankTrans, s.Paxos.acceptNumber, "INTRA")

			s.Bank.UnlockClient(transaction.Sender)
			s.Bank.UnlockClient(transaction.Receiver)
			go s.InitiateRequest()
		}

		// Print balance of x and y
		// log.Println("Balance of x:", s.Bank.GetBalance(s.Paxos.acceptValue.Sender))
		// log.Println("Balance of y:", s.Bank.GetBalance(s.Paxos.acceptValue.Receiver))
	} else {
		if s.Paxos.ballotNumber == s.Paxos.acceptNumber {
			// Only if Paxos hasn't been reset
			log.Println("Unlocking accepted lock for the same Paxos run")
			s.Paxos.accepted.lock.Unlock()
		}
	}

	return nil
}

func (s *Server) sendAbort(transaction rpcmanager.Transaction) {
	log.Println("Sending Abort for: ", transaction)

	message := rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.StatusInterface{
			Transaction: transaction,
			Status:      "ABORT",
		},
	}

	// Send Abort message to client
	s.RPCManager.RPCClients["Client"].Call("Client.TransactionStatus", message, nil)
}

func (s *Server) FinalCrossShardTransactionStatus(message rpcmanager.Message, reply *int) error {

	status := message.Payload.(rpcmanager.StatusInterface).Status
	transaction := message.Payload.(rpcmanager.StatusInterface).Transaction
	log.Println("Received FinalCrossShardTransactionStatus from", message.From, ":", status, transaction)

	bankTrans := bank.Transaction{
		Sender:    transaction.Sender,
		Receiver:  transaction.Receiver,
		Amount:    transaction.Amount,
		ID:        transaction.ID,
		StartTime: transaction.StartTime,
	}

	if status == "COMMIT" {
		// do nothing
		log.Println("Committing transaction to datastore")
		s.Bank.DataStore.AddTransaction(bankTrans, s.Paxos.acceptNumber, "COMMITTED")
		s.Bank.UnlockClient(transaction.Sender)
		s.Bank.UnlockClient(transaction.Receiver)
	} else {
		// Undo transaction. TODO: Add WAL here
		s.Bank.UndoTransaction(transaction.ID)
		s.Bank.DataStore.AddTransaction(bankTrans, s.Paxos.acceptNumber, "ABORTED")
		// newTransaction := rpcmanager.Transaction{
		// 	Sender:    transaction.Receiver,
		// 	Receiver:  transaction.Sender,
		// 	Amount:    transaction.Amount,
		// 	ID:        transaction.ID,
		// 	StartTime: transaction.StartTime,
		// }

		// s.executeTransaction(newTransaction)
	}

	if s.IAmLeader() {
		message.From = s.StringId
		go s.RPCManager.BroadcastMessage("Server.FinalCrossShardTransactionStatus", message, s.ClusterId)
	}
	*reply = 1
	return nil
}

func (s *Server) IAmLeader() bool {
	for _, value := range s.RPCManager.ShardLeaderMapping {
		if value == s.StringId {
			return true
		}
	}
	return false
}
