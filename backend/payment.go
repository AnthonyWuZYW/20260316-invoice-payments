package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// PaymentRecord matches the struct you want to keep
type PaymentRecord struct {
	InvoiceID int     `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	PaidAt    string  `json:"paid_at"`
}

// PaymentDispatcher handles POST /api/invoices/{id}/payments
func paymentDispatcher(w http.ResponseWriter, r *http.Request) {
	// Extract Invoice ID from URL: /api/invoices/{id}/payments
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		sendJSONError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	invoiceID, err := strconv.Atoi(parts[2])
	if err != nil {
		sendJSONError(w, "Invalid Invoice ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming amount
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// BUSINESS RULE: Payment must be positive
	if req.Amount <= 0 {
		sendJSONError(w, "Payment amount must be positive", http.StatusBadRequest)
		return
	}

	// Execute the Transaction Logic
	err = recordPayment(invoiceID, req.Amount)
	if err != nil {
		// Differentiate between Business Rule violations and Server Errors
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "overpayment") || strings.Contains(err.Error(), "status") {
			status = http.StatusBadRequest
		}
		sendJSONError(w, err.Error(), status)
		return
	}

	// 4. Success Response
	sendJSONResponse(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "Payment processed successfully",
	})
}

func recordPayment(invoiceID int, newPaymentAmount float64) error {
	// Start Transaction for Data Integrity and Concurrency
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Defer rollback so it only triggers if we don't hit tx.Commit()
	defer tx.Rollback()

	// Get current Invoice Total and Status
	var invoiceTotal float64
	var status string
	err = tx.QueryRow(`SELECT amount, status FROM invoices WHERE id = ?`, invoiceID).Scan(&invoiceTotal, &status)
	if err != nil {
		return fmt.Errorf("invoice not found")
	}

	// BUSINESS RULE: No payments on PAID or VOID
	if status == "PAID" || status == "VOID" {
		return fmt.Errorf("cannot apply payment to an invoice with status: %s", status)
	}

	// Get sum of existing payments
	var alreadyPaid sql.NullFloat64
	err = tx.QueryRow(`SELECT SUM(amount) FROM payments WHERE invoice_id = ?`, invoiceID).Scan(&alreadyPaid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Validation: Prevent Overpayment
	remainingBalance := invoiceTotal - alreadyPaid.Float64
	// Using a small epsilon  to handle floating point math precision
	if newPaymentAmount > (remainingBalance + 0.0001) {
		return fmt.Errorf("overpayment: remaining balance is only %.2f", remainingBalance)
	}

	// Insert the payment record
	_, err = tx.Exec(`INSERT INTO payments (invoice_id, amount, paid_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, invoiceID, newPaymentAmount)
	if err != nil {
		return err
	}

	// Update Status to 'PAID' if fully paid
	if (remainingBalance - newPaymentAmount) < 0.0001 {
		_, err = tx.Exec(`UPDATE invoices SET status = 'PAID' WHERE id = ?`, invoiceID)
		if err != nil {
			return err
		}
	}

	// Finalize the changes
	return tx.Commit()
}
