package server

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/notoriouscode97/go-microservices/currency/data"
	protos "github.com/notoriouscode97/go-microservices/currency/protos/currency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"time"
)

type Currency struct {
	rates         *data.ExchangeRates
	log           hclog.Logger
	subscriptions map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest
}

func (c *Currency) mustEmbedUnimplementedCurrencyServer() {
	//TODO implement me
}

func NewCurrency(r *data.ExchangeRates, l hclog.Logger) *Currency {
	c := &Currency{log: l, rates: r, subscriptions: make(map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest)}
	go c.handleUpdates()

	return c
}

func (c *Currency) handleUpdates() {
	ru := c.rates.MonitorRates(5 * time.Second)
	for range ru {
		c.log.Info("got updated rates")

		// loop over subscribed clients
		for k, v := range c.subscriptions {

			// loop over subscribed rates
			for _, rr := range v {
				r, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
				if err != nil {
					c.log.Error("unable to get updated rate", "base", rr.GetBase().String(), "destination", rr.GetDestination().String())
				}

				err = k.Send(&protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: r})
				if err != nil {
					c.log.Error("unable to send updated rate", "base", rr.Base.String(), "destination", rr.Destination.String())
				}
			}
		}
	}
}

// GetRate implements the CurrencyServer GetRate method and returns the currency exchange rate
// for the two given currencies.
func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle GetRate", "base", rr.GetBase(), "destination", rr.GetDestination())

	if rr.Base == rr.Destination {
		err := status.Newf(
			codes.InvalidArgument,
			"base currency %s can not be the same as the destination currency %s",
			rr.Base.String(),
			rr.Destination.String(),
		)
		err, wde := err.WithDetails(rr)
		if wde != nil {
			return nil, wde
		}

		return nil, err.Err()
	}

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())

	if err != nil {
		return nil, err
	}

	return &protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: rate}, nil
}

// SubscribeRates implements the gRPC bidirectional streaming method for the server
func (c *Currency) SubscribeRates(src protos.Currency_SubscribeRatesServer) error {

	// handle client messages

	for {
		rr, err := src.Recv() // Recv is a blocking method which returns on client data
		// io.EOF signals that the client has closed the connection
		if err == io.EOF {
			c.log.Info("client has closed connection")
			break
		}
		// any other error means the transport between the server and client is unavailable
		if err != nil {
			c.log.Error("unable to read from client", "error", err)
			return err
		}

		c.log.Info("handle client request", "request", rr)

		rrs, ok := c.subscriptions[src]
		if !ok {
			rrs = []*protos.RateRequest{}
		}

		rrs = append(rrs, rr)
		c.subscriptions[src] = rrs
	}

	return nil
}
