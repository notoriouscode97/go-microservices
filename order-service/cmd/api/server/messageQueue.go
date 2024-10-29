package server

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/notoriouscode97/go-microservices/order-service/internal/data"
	"github.com/notoriouscode97/go-microservices/order-service/internal/mailer"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type MessageQueue struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	l      hclog.Logger
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func NewMessageQueue(url string, logger hclog.Logger, mailer mailer.Mailer) (*MessageQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare an exchange for publishing messages. This exchange will be of type 'topic',
	// allowing for routing messages based on routing keys (patterns). The exchange is durable,
	// meaning it will survive broker restarts
	err = ch.ExchangeDeclare(
		"orders_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("could not declare exchange %w", err)
	}

	return &MessageQueue{conn: conn, ch: ch, l: logger, mailer: mailer}, nil
}

func (r *MessageQueue) Close() error {
	if err := r.ch.Close(); err != nil {
		r.l.Error("Error closing RabbitMQ channel", "error", err)
	}
	return r.conn.Close()
}

func (r *MessageQueue) ConsumeOrders(o *data.OrdersDB) error {
	q, err := r.ch.QueueDeclare(
		"orderQueue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.l.Error("could not declare a queue", "error", err)
		return err
	}

	err = r.ch.QueueBind(
		q.Name,
		"order.created",
		"orders_topic",
		false,
		nil,
	)
	if err != nil {
		r.l.Error("could not bind a queue", "error", err)
		return err
	}

	messages, err := r.ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.l.Error("could not register a consumer", "error", err)
		return err
	}

	for msg := range messages {
		var order data.Order
		if err := json.Unmarshal(msg.Body, &order); err != nil {
			r.l.Error("could not decode order", "error", err)
			continue
		}

		if err := o.InsertOrder(order); err != nil {
			r.l.Error("could not process order", "error", err)
			continue
		} else {
			r.sendConfirmationEmail(order)
			r.l.Info("Order processed and email sent to")
		}
	}

	return nil
}

func (r *MessageQueue) sendConfirmationEmail(order data.Order) {
	r.background(func() {
		payload := map[string]any{
			"items": order.Items,
			"email": order.Email,
		}
		err := r.mailer.Send(order.Email, "order_success.tmpl", payload)
		if err != nil {
			r.l.Error("could not send confirmation email", "error", err)
		}
	})
}
