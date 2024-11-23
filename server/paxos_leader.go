package main

import (
	"fmt"

	rpcmanager "github.com/utkarshsingh99/2phasecommit/rpc"
)

// Contact Server

func (s *Server) IntraShardTransaction(message rpcmanager.Message, reply *int) error {
	fmt.Println("Received IntraShardTransaction from", message.From)
	displayTransaction(message.Payload.(rpcmanager.Transaction))

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
	if s.Paxos.active || len(s.RequestQueue.messages) == 0 {
		return // Already in a request
	}

	s.RequestQueue.lock.Lock()

	// Pop first element from request queue
	s.CurRequest = s.RequestQueue.messages[0]
	s.RequestQueue.messages = s.RequestQueue.messages[1:]

	s.RequestQueue.lock.Unlock()

	// Increase ballot number
	s.Paxos.ballotNumber++

	lastTransId, _ := s.Bank.Log.GetLastTransactionID()

	// Broadcast prepare message to all servers in cluster
	s.RPCManager.BroadcastMessage("Server.Prepare", rpcmanager.Message{
		From: s.StringId,
		Payload: rpcmanager.PrepareInterface{
			BallotNumber:      s.Paxos.ballotNumber,
			LastTransactionId: lastTransId,
		},
	}, s.ClusterId)

	// Add timer to initiateRequest again
	// time.AfterFunc(time.Second, func() {
	// 	if !s.Paxos.promises.acked {
	// 		fmt.Println("Initiating request again...", s.Paxos.ballotNumber)
	// 		s.InitiateRequest()
	// 	}
	// })
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

	s.Paxos.promises.lock.Lock()
	fmt.Println("Received Promise from", message.From)
	s.Paxos.promises.messages = append(s.Paxos.promises.messages, message)

	transaction := s.CurRequest.Payload.(rpcmanager.Transaction)

	sender := transaction.Sender
	receiver := transaction.Receiver

	// >= 1 signifies one other server responded
	if len(s.Paxos.promises.messages) >= 1 && !s.Paxos.promises.acked {

		s.Paxos.promises.acked = true
		// Check if there are no locks on data items x and y
		if s.Bank.IsClientLocked(sender) || s.Bank.IsClientLocked(receiver) {
			// Abort transaction
			return nil
		}

		// Check if the balance of x is at least equal to amt.
		if s.Bank.GetBalance(sender) < transaction.Amount {
			// Abort transaction
			return nil
		}

		// Acquire locks on x and y
		s.Bank.LockClient(sender)
		s.Bank.LockClient(receiver)

		s.Paxos.acceptNumber = s.Paxos.ballotNumber
		s.Paxos.acceptValue = transaction

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

	fmt.Println("Received Accepted from", message.From, " transaction: ", message.Payload.(rpcmanager.AcceptedInterface).AcceptValue)
	s.Paxos.accepted.lock.Lock()
	s.Paxos.accepted.messages = append(s.Paxos.accepted.messages, message)

	if len(s.Paxos.accepted.messages) >= 1 && s.Paxos.acceptValue.Amount > 0 {

		s.Paxos.accepted.acked = true

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

		// Print balance of x and y
		fmt.Println("Balance of x:", s.Bank.GetBalance(s.Paxos.acceptValue.Sender))
		fmt.Println("Balance of y:", s.Bank.GetBalance(s.Paxos.acceptValue.Receiver))
		s.Paxos.accepted.lock.Unlock()

		s.ResetPaxos()
	} else {
		s.Paxos.accepted.lock.Unlock()
	}

	return nil
}
