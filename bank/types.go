package bank

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/lib/pq"
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

type dataStoreEntry struct {
	BallotNumber     float64
	CrossShardStatus string
	Sender           int
	Receiver         int
	Amount           int
	ID               string
}

type DataStore struct {
	DB          *sql.DB
	Entries     []dataStoreEntry
	TableSuffix string
	lock        sync.RWMutex
}

type WriteAheadLog struct {
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
	WAL         *WriteAheadLog
	DataStore   *DataStore
}

type Client struct {
	ID      int
	Name    string
	Balance float64
}

var ClusterDataMap = map[string][]int{
	"C1": {1, 1000},
	"C2": {1001, 2000},
	"C3": {2001, 3000},
}
