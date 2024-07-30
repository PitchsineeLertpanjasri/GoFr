package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"

	"gofr.dev/pkg/gofr"
)

type Customer struct {
	ID   int    `json:"id" db:"id" form:"id"`
	Name string `json:"name" db:"name" form:"name"`
}

func main() {
	// Initialise gofr object
	app := gofr.New()

	// Get port from environment variable
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "9000" // Default port
	}

	app.GET("/redis", func(ctx *gofr.Context) (interface{}, error) {
		// Get the value using the Redis instance
		val, err := ctx.Redis.Get(ctx.Context, "test").Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("Error getting value from Redis: %v", err)
			return nil, err
		}
		return val, nil
	})

	app.POST("/customer/{name}", func(ctx *gofr.Context) (interface{}, error) {
		name := ctx.PathParam("name")
		// Inserting a customer row in database using SQL
		_, err := ctx.SQL.ExecContext(ctx, "INSERT INTO customers (name) VALUES (?)", name)
		if err != nil {
			log.Printf("Error inserting customer into database: %v", err)
			return nil, err
		}
		return "Customer added successfully", nil
	})

	app.GET("/customer", func(ctx *gofr.Context) (interface{}, error) {
		var customers []Customer
		// Getting the customer from the database using SQL
		rows, err := ctx.SQL.QueryContext(ctx, "SELECT * FROM customers")
		if err != nil {
			log.Printf("Error querying customers from database: %v", err)
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var customer Customer
			if err := rows.Scan(&customer.ID, &customer.Name); err != nil {
				log.Printf("Error scanning customer row: %v", err)
				return nil, err
			}
			customers = append(customers, customer)
		}
		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			return nil, err
		}
		// Return the customers
		return customers, nil
	})

	// Create customer
	app.POST("/customer", func(ctx *gofr.Context) (interface{}, error) {
		var customer Customer
		if err := ctx.Bind(&customer); err != nil {
			log.Printf("Error binding customer data: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		_, err := ctx.SQL.ExecContext(ctx.Context, "INSERT INTO customers (name) VALUES (?)", customer.Name)
		if err != nil {
			log.Printf("Error inserting customer into database: %v", err)
			return nil, fmt.Errorf("ErrorEntityAlreadyExist: %v", err)
		}
		return "Customer added successfully", nil
	})

	// Read all customers
	app.GET("/customers", func(ctx *gofr.Context) (interface{}, error) {
		var customers []Customer
		rows, err := ctx.SQL.QueryContext(ctx.Context, "SELECT * FROM customers")
		if err != nil {
			log.Printf("Error querying customers from database: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var customer Customer
			if err := rows.Scan(&customer.ID, &customer.Name); err != nil {
				log.Printf("Error scanning customer row: %v", err)
				return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
			}
			customers = append(customers, customer)
		}
		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		return customers, nil
	})

	// Read single customer
	app.GET("/customer/{id}", func(ctx *gofr.Context) (interface{}, error) {
		id := ctx.PathParam("id")
		var customer Customer
		err := ctx.SQL.QueryRowContext(ctx.Context, "SELECT id, name FROM customers WHERE id = ?", id).Scan(&customer.ID, &customer.Name)
		if err == sql.ErrNoRows {
			log.Printf("Customer not found: %v", err)
			return nil, fmt.Errorf("ErrorEntityNotFound: %v", err)
		} else if err != nil {
			log.Printf("Error querying customer from database: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		return customer, nil
	})

	// Update customer
	app.PUT("/customer/{id}", func(ctx *gofr.Context) (interface{}, error) {
		id := ctx.PathParam("id")
		var customer Customer
		if err := ctx.Bind(&customer); err != nil {
			log.Printf("Error binding customer data: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		log.Printf("Updating customer ID %s with name: %s", id, customer.Name)
		_, err := ctx.SQL.ExecContext(ctx.Context, "UPDATE customers SET name = ? WHERE id = ?", customer.Name, id)
		if err != nil {
			log.Printf("Error updating customer in database: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		return "Customer updated successfully", nil
	})

	// Delete customer
	app.DELETE("/customer/{id}", func(ctx *gofr.Context) (interface{}, error) {
		id := ctx.PathParam("id")
		_, err := ctx.SQL.ExecContext(ctx.Context, "DELETE FROM customers WHERE id = ?", id)
		if err != nil {
			log.Printf("Error deleting customer from database: %v", err)
			return nil, fmt.Errorf("ErrorInvalidParam: %v", err)
		}
		return "Customer deleted successfully", nil
	})

	// Start the application
	log.Printf("Starting server on port: %s", port)
	app.Run()
}
