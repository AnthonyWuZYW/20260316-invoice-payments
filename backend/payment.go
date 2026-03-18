package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type PaymentRecord struct {
	InvoiceID int     `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	PaidAt    string  `json:"paid_at"`
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	// Allow Post
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reveived Payment
	fmt.Printf("Success: Received %f for Invoice %d\n", req.Amount, req.InvoiceID)

	// Send Success response
	sendJSONResponse(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "Payment processed successfully",
		Data:    req,
	})
}

func recordPayment(invoiceID int, newPaymentAmount float64) error {
	// Start Transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Get the Invoice Total and Current Payments Sum
	var invoiceTotal float64
	var alreadyPaid sql.NullFloat64

	// Get total amount owed
	err = tx.QueryRow(`SELECT amount FROM invoices WHERE id = ?`, invoiceID).Scan(&invoiceTotal)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("invoice not found")
	}

	// Get sum of existing payments
	err = tx.QueryRow(`SELECT SUM(amount) FROM payments WHERE invoice_id = ?`, invoiceID).Scan(&alreadyPaid)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	// Payment logic
	remainingBalance := invoiceTotal - alreadyPaid.Float64
	if newPaymentAmount > remainingBalance {
		tx.Rollback()
		return fmt.Errorf("overpayment: remaining balance is only %.2f", remainingBalance)
	}

	// Record the payment since it's valid
	_, err = tx.Exec(`INSERT INTO payments (invoice_id, amount) VALUES (?, ?)`, invoiceID, newPaymentAmount)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update Status to 'PAID' if the balance is now zero
	if newPaymentAmount == remainingBalance {
		_, err = tx.Exec(`UPDATE invoices SET status = 'PAID' WHERE id = ?`, invoiceID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Finalize
	return tx.Commit()
}
