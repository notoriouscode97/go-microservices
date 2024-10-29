package data

import (
	"context"
	"database/sql"
	"time"
)

type Order struct {
	Email string      `json:"email"`
	Items []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrdersDB struct {
	db *sql.DB
}

func NewOrdersDB(db *sql.DB) *OrdersDB {
	return &OrdersDB{db: db}
}

func (o *OrdersDB) InsertOrder(order Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Begin a transaction with context
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var orderID int
	err = tx.QueryRowContext(ctx, "INSERT INTO orders (email, order_date) VALUES ($1, $2) RETURNING id", order.Email, time.Now()).Scan(&orderID)
	if err != nil {
		rError := tx.Rollback()
		if rError != nil {
			return rError
		}
		return err
	}

	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, "INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)", orderID, item.ProductID, item.Quantity)
		if err != nil {
			rError := tx.Rollback()
			if rError != nil {
				return rError
			}
			return err
		}
	}

	return tx.Commit()
}
