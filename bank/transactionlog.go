package bank

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func NewLog() *Log {
	return &Log{
		Transactions: make([]Transaction, 0),
		Locks:        sync.RWMutex{},
	}
}

func (t *Log) AddTransaction(transaction Transaction) {
	t.Locks.Lock()
	transaction.LogTime = time.Now()
	t.Transactions = append(t.Transactions, transaction)
	t.Locks.Unlock()
}

func (t *Log) GetLastTransactionID() (string, error) {
	t.Locks.RLock()
	defer t.Locks.RUnlock()
	if len(t.Transactions) == 0 {
		return "", errors.New("Transaction log is empty")
	}
	return t.Transactions[len(t.Transactions)-1].ID, nil
}

func (t *Log) CompareMyLog(id1 string) (bool, string) {
	t.Locks.RLock()
	defer t.Locks.RUnlock()

	lastID, _ := t.GetLastTransactionID()
	if lastID == id1 {
		return true, "all_updated"
	} else {
		for _, transaction := range t.Transactions[:len(t.Transactions)-1] {
			if transaction.ID == id1 {
				return false, "they_outdated"
			}
		}
		if len(t.Transactions) > 0 && id1 == "" {
			return false, "they_outdated"
		}
		return false, "me_outdated"
	}
}

func (t *Log) UpdateLog(transactions []Transaction) {
	// Add new transactions to the log
	for _, transaction := range transactions {
		t.AddTransaction(transaction)
	}
}

func (t *Log) SendRemainingLog(transactionID string) []Transaction {
	t.Locks.RLock()
	defer t.Locks.RUnlock()
	index := 0
	fmt.Println("SendRemiainingLog Transactions: ", t.Transactions)
	for _, transaction := range t.Transactions {
		index += 1
		if transaction.ID == transactionID {
			// Return all remaining transactions after the current transaction
			return t.Transactions[index:]
		}
	}
	return []Transaction{}
}
