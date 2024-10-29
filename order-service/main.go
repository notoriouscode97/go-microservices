package main

import (
	"context"
	"database/sql"
	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
	"github.com/notoriouscode97/go-microservices/order-service/cmd/api/config"
	"github.com/notoriouscode97/go-microservices/order-service/cmd/api/server"
	"github.com/notoriouscode97/go-microservices/order-service/internal/data"
	"github.com/notoriouscode97/go-microservices/order-service/internal/mailer"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// load .env variables
	cfg, err := config.LoadConfig()

	// init logger
	l := hclog.Default()

	if err != nil {
		l.Error("loading env config", err)
		os.Exit(1)
	}

	// open database connection
	openDb, err := openDB(*cfg)

	if err != nil {
		l.Error("error opening db connection: %s\n", err)
		os.Exit(1)
	}

	// create database instance
	db := data.NewOrdersDB(openDb)

	// Establish connection to RabbitMQ
	rabbit, err := server.NewMessageQueue(
		"amqp://guest:guest@localhost:5672/",
		l,
		mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
	)
	if err != nil {
		l.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}

	defer func(rabbit *server.MessageQueue) {
		err := rabbit.Close()
		if err != nil {
			l.Error("Failed to close RabbitMQ connection", "error", err)
		}
	}(rabbit)

	// Start consuming orders
	go func() {
		if err := rabbit.ConsumeOrders(db); err != nil {
			log.Fatalf("Error consuming orders: %v", err)
		}
	}()

	log.Println("Order processing service running...")

	// Block the main goroutine
	waitForShutdown(rabbit)
}

// waitForShutdown blocks until a termination signal is received
func waitForShutdown(rabbit *server.MessageQueue) {
	// Create a channel to listen for OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigs
	log.Println("Shutdown signal received. Initiating graceful shutdown...")

	// Close the RabbitMQ connection and channel
	if err := rabbit.Close(); err != nil {
		log.Printf("Error closing RabbitMQ connection: %v", err)
	}

	// Here you can add any additional cleanup actions if needed,
	// like waiting for active workers to finish processing.

	// You can also add a timeout to wait for workers to finish processing
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Cleanup completed. Exiting.")
}

func openDB(cfg config.Settings) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Db.Dsn)

	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.Db.MaxOpenConnections)

	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.Db.MaxIdleConnections)

	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.Db.MaxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
