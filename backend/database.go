package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Initialize Database
func initDB() *sql.DB {
	var err error
	db, err = sql.Open("sqlite3", "./ecapital.db")
	if err != nil {
		log.Fatal("Could not connect to SQLite:", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal("Could not enable foreign keys:", err)
	}

	schema := `
        CREATE TABLE IF NOT EXISTS customers (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE
        );

        CREATE TABLE IF NOT EXISTS invoices (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            customer_id INTEGER NOT NULL,
            amount DECIMAL(10,2) NOT NULL,
            currency TEXT DEFAULT 'USD',
            status TEXT CHECK(status IN ('DRAFT', 'PENDING', 'PAID', 'VOID')) DEFAULT 'PENDING',
            issued_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            due_at DATETIME,
            FOREIGN KEY (customer_id) REFERENCES customers(id)
        );

        CREATE TABLE IF NOT EXISTS payments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            invoice_id INTEGER NOT NULL,
            amount DECIMAL(10,2) NOT NULL,
            paid_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (invoice_id) REFERENCES invoices(id)
        );`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal("Table creation failed:", err)
	}
	return db
}

// Seed Structures
type SeedData struct {
	Customers []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"customers"`
	Invoices []struct {
		ID         int     `json:"id"`
		CustomerID int     `json:"customer_id"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		IssuedAt   string  `json:"issued_at"`
		DueAt      string  `json:"due_at"`
		Status     string  `json:"status"`
	} `json:"invoices"`
	Payments []struct {
		ID        int     `json:"id"`
		InvoiceID int     `json:"invoice_id"`
		Amount    float64 `json:"amount"`
		PaidAt    string  `json:"paid_at"`
	} `json:"payments"`
}

// SeedDatabase fills the DB with initial data if it's empty
func SeedDatabase(jsonData string) error {
	var seed SeedData
	if err := json.Unmarshal([]byte(jsonData), &seed); err != nil {
		return err
	}

	// Check if already seeded
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already has data
	}

	fmt.Println("Seeding database with initial eCapital data...")

	// Insert Customers
	for _, c := range seed.Customers {
		_, err := db.Exec("INSERT INTO customers (id, name) VALUES (?, ?)", c.ID, c.Name)
		if err != nil {
			return err
		}
	}

	// Insert Invoices
	for _, i := range seed.Invoices {
		_, err := db.Exec(`INSERT INTO invoices (id, customer_id, amount, currency, issued_at, due_at, status) 
            VALUES (?, ?, ?, ?, ?, ?, ?)`, i.ID, i.CustomerID, i.Amount, i.Currency, i.IssuedAt, i.DueAt, i.Status)
		if err != nil {
			return err
		}
	}

	// Insert Payments
	for _, p := range seed.Payments {
		_, err := db.Exec("INSERT INTO payments (id, invoice_id, amount, paid_at) VALUES (?, ?, ?, ?)",
			p.ID, p.InvoiceID, p.Amount, p.PaidAt)
		if err != nil {
			return err
		}
	}

	fmt.Println("Seeding successful!")
	return nil
}
