package bank

import (
	"fmt"
	"sync"
)

func New(startId int, endId int) *Bank {
	balance := &BankBalance{
		Clients:     make(map[int]int),
		Locks:       make(map[int]*sync.RWMutex),
		LockTracker: make(map[int]bool),
	}

	// Prepopulate the Locks map
	for i := startId; i <= endId; i++ {
		balance.Locks[i] = &sync.RWMutex{}
		balance.LockTracker[i] = false
	}

	bankPointer := &Bank{
		BankBalance: balance,
		Log:         NewLog(),
	}
	bankPointer.InitiateClients(startId, endId)

	return bankPointer
}

func (b *Bank) InitiateClients(startId int, endId int) {
	for i := startId; i <= endId; i++ {
		b.LockClient(i)
		b.BankBalance.Clients[i] = 10
		b.UnlockClient(i)
	}
}

func (b *Bank) GetBalance(clientId int) int {
	// Check if client exists
	if _, ok := b.BankBalance.Clients[clientId]; !ok {
		return -1
	}

	return b.BankBalance.Clients[clientId]
}

func (b *Bank) UpdateBalance(clientId int, amount int) {
	b.BankBalance.Clients[clientId] += amount
}

func (b *Bank) LockClient(clientId int) {
	b.BankBalance.Locks[clientId].Lock()
	b.BankBalance.LockTracker[clientId] = true
}
func (b *Bank) UnlockClient(clientId int) {
	b.BankBalance.Locks[clientId].Unlock()
	b.BankBalance.LockTracker[clientId] = false
}

func (b *Bank) IsClientLocked(clientId int) bool {
	return b.BankBalance.LockTracker[clientId]
}

func (b *Bank) CommitTransaction(transaction Transaction) {
	fmt.Println("Actually committing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)
	b.Log.AddTransaction(transaction)

	b.BankBalance.Clients[transaction.Sender] -= transaction.Amount
	b.BankBalance.Clients[transaction.Receiver] += transaction.Amount

}
