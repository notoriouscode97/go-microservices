package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/data"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/http"
	"time"
)

var (
	conn *amqp.Connection
	ch   *amqp.Channel
)

// KeyOrder is a key used for the OrderRequest object in the context
type KeyOrder struct{}

// swagger:route POST /orders orders createOrder
// Create a new order
//
// responses:
//	201: noContentResponse
//  422: errorValidation
//  500: errorResponse

// CreateOrder HTTP handler for creating orders
func (p *Products) CreateOrder(rw http.ResponseWriter, r *http.Request) {
	orderReq := r.Context().Value(KeyOrder{}).(*data.OrderRequest)

	v := data.NewValidation()

	// Validate the product
	if err := v.Validate(orderReq); err != nil { // Implement this validation function
		p.l.Error("Validation error:", err)
		rw.WriteHeader(http.StatusUnprocessableEntity) // 422
		errString := fmt.Sprintf("Validation error: %s", err.Errors())
		_ = data.ToJSON(&GenericError{Message: errString}, rw)
		return
	}

	// Publish the order
	err := p.publishOrderToQueue(orderReq)
	if err != nil {
		p.l.Error("error publishing order:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		_ = data.ToJSON(&GenericError{Message: "could not process order"}, rw)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	_ = data.ToJSON(&GenericResponse{Message: "order created"}, rw)
}

// Connect to RabbitMQ and publish a message
func (p *Products) publishOrderToQueue(order *data.OrderRequest) error {
	orderBytes, err := json.Marshal(order)
	if err != nil {
		p.l.Error("error serializing order: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"orders_topic",  // Exchange
		"order.created", // Routing key
		false,           // Mandatory
		false,           // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        orderBytes,
		},
	)

	if err != nil {
		p.l.Error("error publishing order to RabbitMQ: %v", err)
		return err
	}

	return nil
}

func (p *Products) InitRabbitMQ() error {
	var err error
	conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	p.l.Info("Connected to RabbitMQ") // Log connection success

	ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	p.l.Info("Opened RabbitMQ channel") // Log channel open success

	// Declare the exchange once
	err = ch.ExchangeDeclare(
		"orders_topic", // Exchange name
		"topic",        // Exchange type
		true,           // Durable
		false,          // Auto-deleted
		false,          // Internal
		false,          // No-wait
		nil,            // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.l.Info("Declared RabbitMQ exchange: orders") // Log exchange declaration success

	return nil
}

// ShutdownRabbitMQ closes the RabbitMQ channel and connection.
func (p *Products) ShutdownRabbitMQ() {
	if ch != nil {
		if err := ch.Close(); err != nil {
			p.l.Error("error closing RabbitMQ channel: %v", err)
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			p.l.Error("error closing RabbitMQ connection: %v", err)
		}
	}
}
