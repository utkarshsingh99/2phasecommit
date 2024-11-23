package bank

import (
	"sync"
	"time"
)

type Transaction struct {
	Sender    int
	Receiver  int
	Amount    int
	ID        string
	StartTime time.Time
	LogTime   time.Time
}

type Log struct {
	Transactions []Transaction
	Locks        sync.RWMutex
}

type BankBalance struct {
	Locks       map[int]*sync.RWMutex
	LockTracker map[int]bool
	Clients     map[int]int
}

type Bank struct {
	BankBalance *BankBalance
	Log         *Log
}

var ClusterDataMap = map[string][]int{
	"C1": {1, 1000},
	"C2": {1001, 2000},
	"C3": {2001, 3000},
}
