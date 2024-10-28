package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	protos "github.com/notoriouscode97/go-microservices/currency/protos/currency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"github.com/hashicorp/go-hclog"
)

// ErrProductNotFound is an error raised when a product can not be found in the database
var ErrProductNotFound = fmt.Errorf("product not found")

// Product defines the structure for an API product
// swagger:model
type Product struct {
	// the id for the product
	//
	// required: false
	// min: 1
	ID int `json:"id"` // Unique identifier for the product

	// the name for this product
	//
	// required: true
	// max length: 255
	Name string `json:"name" validate:"required"`

	// the description for this product
	//
	// required: false
	// max length: 10000
	Description string `json:"description"`

	// the price for the product
	//
	// required: true
	// min: 0.01
	Price float64 `json:"price" validate:"required,gt=0"`

	// the SKU for the product
	//
	// required: true
	// pattern: [a-z]+-[a-z]+-[a-z]+
	SKU string `json:"sku" validate:"sku"`
}

// Products defines a slice of Product
type Products []*Product

type ProductsDB struct {
	db       *sql.DB
	currency protos.CurrencyClient
	log      hclog.Logger
	rates    map[string]float64
	client   protos.Currency_SubscribeRatesClient
}

func NewProductsDB(db *sql.DB, c protos.CurrencyClient, l hclog.Logger) *ProductsDB {
	pb := &ProductsDB{db, c, l, make(map[string]float64), nil}
	go pb.handleUpdates()

	return pb
}

func (p *ProductsDB) handleUpdates() {
	sub, err := p.currency.SubscribeRates(context.Background())
	if err != nil {
		p.log.Error("Unable to subscribe for rates", "error", err)
		return
	}

	p.client = sub

	for {
		rr, err := sub.Recv()
		p.log.Info("received updated rate from server", "dest", rr.GetDestination().String())
		rr.GetRate()
		if err != nil {
			p.log.Error("error receiving message", "error", err)
			return
		}

		p.rates[rr.Destination.String()] = rr.Rate
	}
}

// GetProducts returns all products from the database
func (p *ProductsDB) GetProducts(currency string) (Products, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, "SELECT id, name, description, price, sku FROM products")

	if err != nil {
		p.log.Error("error fetching products:", err)
		return nil, err
	}

	defer rows.Close()

	var products Products

	var rate = 1.0 // Default rate if no currency is provided

	if currency != "" {
		rate, err = p.getRate(currency)
		if err != nil {
			p.log.Error("unable to get rate", "currency", currency, "error", err)
			return nil, err
		}
	}

	for rows.Next() {
		var prod Product

		if err := rows.Scan(&prod.ID, &prod.Name, &prod.Description, &prod.Price, &prod.SKU); err != nil {
			p.log.Error("error scanning product", "error", err)
			return nil, err
		}

		prod.Price *= rate
		products = append(products, &prod)
	}

	if err = rows.Err(); err != nil {
		p.log.Error("error fetching products:", err)
		return nil, err
	}

	return products, nil
}

// GetProductByID returns a single product which matches the id from the
// database.
// If a product is not found this function returns a ProductNotFound error
func (p *ProductsDB) GetProductByID(id int, currency string) (*Product, error) {
	if id < 1 {
		return nil, ErrProductNotFound
	}

	query := `SELECT id, name, description, price, sku FROM products
	WHERE id = $1`

	var product Product

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrProductNotFound
		default:
			return nil, err
		}
	}

	var rate = 1.0 // Default rate if no currency is provided

	if currency != "" {
		rate, err = p.getRate(currency)
		if err != nil {
			p.log.Error("unable to get rate", "currency", currency, "error", err)
			return nil, err
		}
	}

	product.Price *= rate

	return &product, nil
}

// UpdateProduct replaces a product in the database with the given
// item.
// If a product with the given id does not exist in the database
// this function returns a ProductNotFound error
func (p *ProductsDB) UpdateProduct(pr *Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, sku = $4 WHERE id = $5`

	args := []any{
		pr.Name,
		pr.Description,
		pr.Price,
		pr.SKU,
		pr.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProductNotFound // No rows updated
	}

	return nil
}

// AddProduct adds a new product to the database
func (p *ProductsDB) AddProduct(pr *Product) error {
	query := `INSERT INTO products (name, description, price, sku) VALUES ($1, $2, $3, $4) RETURNING id`

	args := []any{
		pr.Name,
		pr.Description,
		pr.Price,
		pr.SKU,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, args...).Scan(&pr.ID)

	if err != nil {
		return err
	}

	return nil
}

// DeleteProduct deletes a product from the database
func (p *ProductsDB) DeleteProduct(id int) error {
	if id < 1 {
		return ErrProductNotFound
	}

	query := `DELETE FROM products WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (p *ProductsDB) getRate(destination string) (float64, error) {
	// if cached return
	if r, ok := p.rates[destination]; ok {
		return r, nil
	}

	rr := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value[destination]),
	}

	// get initial rate
	resp, err := p.currency.GetRate(context.Background(), rr)

	if err != nil {
		if s, ok := status.FromError(err); ok {
			md := s.Details()[0].(*protos.RateRequest)

			if s.Code() == codes.InvalidArgument {
				return -1, fmt.Errorf("unable to get rate from currency server, destination and base currencies can not be the same, base: %s, dest: %s", md.Base.String(), md.Destination.String())
			}
			return -1, fmt.Errorf("unable to get rate from currency server, base: %s, dest: %s", md.Base.String(), md.Destination.String())
		}

		return -1, err
	}

	// update cache
	p.rates[destination] = resp.Rate // update cache

	// subscribe for updates
	_ = p.client.Send(rr)

	return resp.Rate, err
}
