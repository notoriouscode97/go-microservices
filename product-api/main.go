package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
	protos "github.com/notoriouscode97/go-microservices/currency/protos/currency"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/config"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/data"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
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

	v := data.NewValidation()

	conn, err := grpc.NewClient("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("error creating gRPC client", err)
		os.Exit(1)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			l.Error("error closing gRPC client connection", err)
		}
	}(conn)

	// create client
	cc := protos.NewCurrencyClient(conn)

	openDb, err := openDB(*cfg)

	if err != nil {
		l.Error("error opening db connection: %s\n", err)
		os.Exit(1)
	}

	// create database instance
	db := data.NewProductsDB(openDb, cc, l)

	// create the server
	ph := handlers.NewProducts(l, v, db)

	// init message broker
	err = ph.InitRabbitMQ()

	if err != nil {
		l.Error("error initializing RabbitMQ", err)
		os.Exit(1)
	}

	// create a new serve mux and register the server
	sm := mux.NewRouter()

	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll).Queries("currency", "{[A-Z]{3}}")
	getR.HandleFunc("/products", ph.ListAll)

	getR.HandleFunc("/products/{id:[0-9]+}", ph.ListSingle).Queries("currency", "{[A-Z]{3}}")
	getR.HandleFunc("/products/{id:[0-9]+}", ph.ListSingle)

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/products", ph.Update)
	putR.Use(ph.MiddlewareValidateProduct)

	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/products", ph.Create)
	postR.Use(ph.MiddlewareValidateProduct)

	deleteR := sm.Methods(http.MethodDelete).Subrouter()
	deleteR.HandleFunc("/products/{id:[0-9]+}", ph.Delete)

	// handler for documentation
	ops := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(ops, nil)

	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	rabbitR := sm.Methods(http.MethodPost).Subrouter()
	rabbitR.HandleFunc("/orders", ph.CreateOrder)
	rabbitR.Use(ph.MiddlewareValidateOrder)

	// CORS
	ch := gohandlers.CORS(gohandlers.AllowedOrigins([]string{cfg.CorsAllowedOrigin}))

	// create a new server
	s := &http.Server{
		Addr:         cfg.BindAddress,                                  // configure the bind address
		Handler:      ch(sm),                                           // set the default handler
		ErrorLog:     l.StandardLogger(&hclog.StandardLoggerOptions{}), // set the logger for the server
		ReadTimeout:  5 * time.Second,                                  // max time to read request from the client
		WriteTimeout: 10 * time.Second,                                 // max time to write response to the client
		IdleTimeout:  120 * time.Second,                                // max time for connections using TCP Keep-Alive
	}

	// start the server
	go func() {
		l.Info(fmt.Sprintf("Starting server on %s", cfg.BindAddress))

		err := s.ListenAndServe()

		if err != nil {
			l.Error("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interrupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	l.Info("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	err = s.Shutdown(ctx)

	if err != nil {
		l.Error("Error starting server: %s\n", err)
		os.Exit(1)
	}

	// Call ShutdownRabbitMQ to close RabbitMQ connections after the server is shut down
	ph.ShutdownRabbitMQ()
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
