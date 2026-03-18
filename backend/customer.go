package main

import "database/sql"

// Insert a new Customer
func createCustomer(name string) (int64, error) {
	var id int64

	// Check if the customer exist
	query := `SELECT id FROM customers WHERE name = ?`
	err := db.QueryRow(query, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Customer not found create customer
	insertQuery := `INSERT INTO customers (name) VALUES (?)`
	result, err := db.Exec(insertQuery, name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
