package bank

import (
	"fmt"
	"sync"
)

func New(startId int, endId int, tableSuffix string) *Bank {
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
		WAL: &WriteAheadLog{
			Transactions: make([]Transaction, 0),
			Locks:        sync.RWMutex{},
		},
		DataStore: &DataStore{
			Entries:     make([]dataStoreEntry, 0),
			lock:        sync.RWMutex{},
			TableSuffix: tableSuffix,
		},
	}
	bankPointer.DataStore.InitializeSQL("postgres://newuser:123456@localhost:5432/bank?sslmode=disable")

	bankPointer.InitiateClients(startId, endId)

	return bankPointer
}

func (b *Bank) InitiateClients(startId int, endId int) {
	for i := startId; i <= endId; i++ {
		b.LockClient(i)
		b.DataStore.AddClient(fmt.Sprintf("%d", i), 10)
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
	if _, ok := b.BankBalance.Locks[clientId]; !ok {
		return
	}
	b.BankBalance.Locks[clientId].Lock()
	b.BankBalance.LockTracker[clientId] = true
}
func (b *Bank) UnlockClient(clientId int) {
	if _, ok := b.BankBalance.Locks[clientId]; !ok {
		return
	}
	b.BankBalance.Locks[clientId].Unlock()
	b.BankBalance.LockTracker[clientId] = false
}

func (b *Bank) IsClientLocked(clientId int) bool {
	if _, ok := b.BankBalance.LockTracker[clientId]; !ok {
		return false
	}
	return b.BankBalance.LockTracker[clientId]
}

func (b *Bank) CommitTransaction(transaction Transaction) {
	// fmt.Println("Actually committing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)
	b.Log.AddTransaction(transaction)

	if _, ok := b.BankBalance.Clients[transaction.Sender]; ok {
		b.BankBalance.Clients[transaction.Sender] -= transaction.Amount
		b.DataStore.UpdateClientBalance(transaction.Sender, b.BankBalance.Clients[transaction.Sender])
	}

	if _, ok := b.BankBalance.Clients[transaction.Receiver]; ok {
		b.BankBalance.Clients[transaction.Receiver] += transaction.Amount
		b.DataStore.UpdateClientBalance(transaction.Receiver, b.BankBalance.Clients[transaction.Receiver])
	}
}
