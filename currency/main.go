package main

import (
	"github.com/hashicorp/go-hclog"
	protos "github.com/notoriouscode97/go-microservices/currency/protos/currency"
	"github.com/notoriouscode97/go-microservices/currency/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
)

func main() {
	log := hclog.Default()

	gs := grpc.NewServer()
	c := server.NewCurrency(log)

	protos.RegisterCurrencyServer(gs, c)

	reflection.Register(gs)

	l, err := net.Listen("tcp", ":9092")
	if err != nil {
		log.Error("Listen error", "error", err)
		os.Exit(1)
	}

	gs.Serve(l)
}
