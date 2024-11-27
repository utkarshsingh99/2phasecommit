package bank

import "log"

func (w *WriteAheadLog) Append(transaction Transaction) {
	w.Locks.Lock()
	log.Println("Adding to log: ", transaction.Sender, transaction.Receiver, transaction.Amount)
	w.Transactions = append(w.Transactions, transaction)
	w.Locks.Unlock()
}

func (b *Bank) UndoTransaction(transactionID string) {
	log.Println("UndoTransaction: ", transactionID)
	w := b.WAL

	w.Locks.Lock()
	defer w.Locks.Unlock()

	for _, transaction := range w.Transactions {
		if transaction.ID == transactionID {
			newTransaction := Transaction{
				Sender:    transaction.Receiver,
				Receiver:  transaction.Sender,
				Amount:    transaction.Amount,
				ID:        transaction.ID,
				StartTime: transaction.StartTime,
			}
			log.Println("Undoing transaction: ", transaction.Sender, transaction.Receiver, transaction.Amount)
			b.CommitTransaction(newTransaction)
			// UNCOMMENT TO MAKE 10th Test case work. But will fail 1st case
			b.UnlockClient(transaction.Receiver)
			b.UnlockClient(transaction.Sender)
			return
		}
	}
	log.Println("Log to undo not found!! This server caused the abort!")
}
