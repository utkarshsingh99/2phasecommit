package bank

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func (d *DataStore) AddTransaction(transaction Transaction, ballotNumber float64, crossShardStatus string) {
	prevBallotNumber := ballotNumber
	if crossShardStatus == "COMMITTED" || crossShardStatus == "ABORTED" {
		log.Println("Updating commit transaction to datastore: ", transaction.Sender, transaction.Receiver, transaction.Amount, transaction.ID)
		for each := range d.Entries {
			log.Println("ID: ", d.Entries[each])
			if d.Entries[each].ID == transaction.ID {
				oldEntry := d.Entries[each]
				prevBallotNumber = d.Entries[each].BallotNumber
				d.Entries[each] = dataStoreEntry{
					BallotNumber:     prevBallotNumber,
					CrossShardStatus: oldEntry.CrossShardStatus,
					Sender:           transaction.Sender,
					Receiver:         transaction.Receiver,
					Amount:           oldEntry.Amount,
					ID:               oldEntry.ID,
				}

				if d.DB != nil {
					tableName := fmt.Sprintf("transactions_%s", d.TableSuffix)
					// Update existing transaction
					query := `UPDATE ? SET ballot_number = ?, cross_shard_status = ?, sender = ?, receiver = ?, amount = ? WHERE id = ?`
					_, err := d.DB.Exec(query, tableName, ballotNumber, crossShardStatus, transaction.Sender, transaction.Receiver, transaction.Amount, transaction.ID)
					if err != nil {
						log.Println("Failed to update transaction in PostgreSQL:", err)
					}
				}
				break
			}
		}
	} else {
		log.Println("Adding prepared transaction to datastore: ", transaction.Sender, transaction.Receiver, transaction.Amount, transaction.ID)
	}
	d.Entries = append(d.Entries, dataStoreEntry{
		BallotNumber:     prevBallotNumber,
		CrossShardStatus: crossShardStatus,
		Sender:           transaction.Sender,
		Receiver:         transaction.Receiver,
		Amount:           transaction.Amount,
		ID:               transaction.ID,
	})

	// Save to PostgreSQL database
	log.Println("Saving transaction to PostgreSQL...", d.DB)
	if d.DB != nil {
		tableName := fmt.Sprintf("transactions_%s", d.TableSuffix)
		query := fmt.Sprintf(`
        INSERT INTO %s (id, ballot_number, cross_shard_status, sender, receiver, amount)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE
        SET ballot_number = EXCLUDED.ballot_number,
            cross_shard_status = EXCLUDED.cross_shard_status,
            sender = EXCLUDED.sender,
            receiver = EXCLUDED.receiver,
            amount = EXCLUDED.amount`, tableName)

		if crossShardStatus == "ABORTED" {
			_, err := d.DB.Exec(query, transaction.ID, ballotNumber, crossShardStatus, transaction.Receiver, transaction.Sender, transaction.Amount)
			if err != nil {
				log.Println("Failed to save transaction to database:", err)
			}
		} else {
			_, err := d.DB.Exec(query, transaction.ID, ballotNumber, crossShardStatus, transaction.Sender, transaction.Receiver, transaction.Amount)
			if err != nil {
				log.Println("Failed to save transaction to database:", err)
			}
		}
	}
}

func (d *DataStore) InitializeSQL(dsn string) error {
	// Example DSN: "postgres://bank_user:password@localhost:5432/bank?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Println("Failed to connect to PostgreSQL database:", err)
		return err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Println("Failed to ping PostgreSQL database:", err)
		return err
	}

	d.DB = db
	log.Println("PostgreSQL database connection established.")

	// Create transactions table
	if err := d.CreateTransactionsTable(); err != nil {
		log.Println("Failed to create transactions table:", err)
		return err
	}

	// Load existing transactions from PostgreSQL
	err = d.LoadFromSQL()
	if err != nil {
		log.Fatal("Failed to load transactions from PostgreSQL:", err)
	}
	return nil
}

func (d *DataStore) CreateTransactionsTable() error {
	tableName := fmt.Sprintf("transactions_%s", d.TableSuffix)
	clientTableName := fmt.Sprintf("clients_%s", d.TableSuffix)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			ballot_number FLOAT,
			cross_shard_status VARCHAR(20),
			sender VARCHAR(255),
			receiver VARCHAR(255),
			amount FLOAT
		)`, tableName)

	query1 := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,        -- Unique ID for each client
    name VARCHAR(255) NOT NULL,  -- Name of the client
    balance FLOAT NOT NULL
	)`, clientTableName)

	_, err := d.DB.Exec(query)
	_, err1 := d.DB.Exec(query1)
	if err != nil || err1 != nil {
		return err
	}
	log.Printf("Created table: %s", tableName)
	return nil
}

func (d *DataStore) LoadFromSQL() error {
	tableName := fmt.Sprintf("transactions_%s", d.TableSuffix)
	query := fmt.Sprintf(`SELECT id, ballot_number, cross_shard_status, sender, receiver, amount FROM %s`, tableName)

	rows, err := d.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entry dataStoreEntry
		err := rows.Scan(&entry.ID, &entry.BallotNumber, &entry.CrossShardStatus, &entry.Sender, &entry.Receiver, &entry.Amount)
		if err != nil {
			return err
		}
		d.Entries = append(d.Entries, entry)
	}

	log.Println("Loaded transactions from SQL into memory.")
	return nil
}

func (d *DataStore) AddClient(name string, initialBalance float64) error {
	tableName := fmt.Sprintf("clients_%s", d.TableSuffix)
	query := fmt.Sprintf(`INSERT INTO %s (name, balance) VALUES ($1, $2)`, tableName)
	_, err := d.DB.Exec(query, name, initialBalance)
	if err != nil {
		log.Println("Failed to add client:", err)
		return err
	}
	log.Printf("Added client: %s with initial balance: %.2f", name, initialBalance)
	return nil
}

func (d *DataStore) UpdateClientBalance(clientID int, newBalance int) error {
	tableName := fmt.Sprintf("clients_%s", d.TableSuffix)
	query := fmt.Sprintf(`UPDATE %s SET balance = $1 WHERE id = $2`, tableName)

	_, err := d.DB.Exec(query, newBalance, clientID)
	if err != nil {
		log.Println("Failed to update client balance:", err)
		return err
	}

	log.Printf("Updated balance for client ID %d to %d", clientID, newBalance)
	return nil
}

func (d *DataStore) GetClientBalance(clientID int) (int, error) {
	tableName := fmt.Sprintf("clients_%s", d.TableSuffix)
	query := fmt.Sprintf(`SELECT balance FROM %s WHERE id = $1`, tableName)

	var balance int
	err := d.DB.QueryRow(query, clientID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
