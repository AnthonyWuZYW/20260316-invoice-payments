package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type InvoiceRequest struct {
	CustomerID int     `json:"customer_id"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	IssueDate  string  `json:"issued_at"`
	DueDate    string  `json:"due_at"`
	Status     string  `json:"status"`
}

type InvoiceRecord struct {
	ID         int     `json:"id"`
	CustomerID int     `json:"customer_id"`
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	IssueDate  string  `json:"issued_at"`
	DueDate    string  `json:"due_at"`
	Status     string  `json:"status"`
}

type InvoiceDetail struct {
	InvoiceRecord                 // Reuse your existing struct fields
	Payments      []PaymentRecord `json:"payments"`
}

func InvoicesDispatcher(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateInvoice(w, r)
	case http.MethodGet:
		handleGetInvoices(w, r)
	default:
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateInvoice handles POST /api/invoices
func handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req InvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create Cusomter
	customerID, err := createCustomer(req.Name)
	if err != nil {
		sendJSONError(w, "Failed to process customer", http.StatusInternalServerError)
		return
	}

	// Create the Invoice
	invoiceID, err := createInvoice(int(customerID), req.Amount, req.Currency, req.IssueDate, req.DueDate)
	if err != nil {
		sendJSONError(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}

	// Send Success response
	sendJSONResponse(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "Invoice Registered successfully",
		Data:    map[string]interface{}{"invoice_id": invoiceID, "customer_id": customerID},
	})
}

// handleGetInvoices handles GET /api/invoices and GET /api/invoices/{id}
func handleGetInvoices(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL (e.g., "/api/invoices/1")
	// We strip the prefix to see if there is an ID at the end
	idStr := r.URL.Path[len("/api/invoices/"):]

	// ID provided: Search by ID
	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "Invalid Invoice ID format", http.StatusBadRequest)
		return
	}

	detail, err := getInvoiceByID(id)
	if err != nil {
		sendJSONError(w, "Invoice not found", http.StatusNotFound)
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{Status: "success", Data: detail})
}

// Insert a new Invoice
func createInvoice(customerID int, amount float64, currency string, issueAt string, dueAt string) (int64, error) {
	query := `INSERT INTO invoices (customer_id, amount, currency, issue_at, due_at, status) 
              VALUES (?, ?, ?, ?, ?, 'PENDING')`
	result, err := db.Exec(query, customerID, amount, currency, issueAt, dueAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func getInvoiceByID(id int) (*InvoiceDetail, error) {
	var detail InvoiceDetail
	detail.Payments = []PaymentRecord{}

	// Get the Core Invoice Details + Customer Name
	query := `
        SELECT i.id, c.name, i.amount, i.currency, i.status, i.issued_at, i.due_at
        FROM invoices i
        JOIN customers c ON i.customer_id = c.id
        WHERE i.id = ?`

	err := db.QueryRow(query, id).Scan(
		&detail.ID,
		&detail.Name,
		&detail.Amount,
		&detail.Currency,
		&detail.Status,
		&detail.IssueDate,
		&detail.DueDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}

		return nil, err
	}

	// 2. Get the Payment Breakdown
	paymentQuery := `SELECT amount, paid_at FROM payments WHERE invoice_id = ? ORDER BY paid_at DESC`
	rows, err := db.Query(paymentQuery, id)
	if err != nil {
		return &detail, nil
	}
	defer rows.Close()

	for rows.Next() {
		var p PaymentRecord
		if err := rows.Scan(&p.Amount, &p.PaidAt); err == nil {
			detail.Payments = append(detail.Payments, p)
		}
	}

	return &detail, nil
}
