package server

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/notoriouscode97/go-microservices/currency/data"
	protos "github.com/notoriouscode97/go-microservices/currency/protos/currency"
)

type Currency struct {
	rates *data.ExchangeRates
	log   hclog.Logger
}

func (c *Currency) mustEmbedUnimplementedCurrencyServer() {
	//TODO implement me
}

func NewCurrency(r *data.ExchangeRates, l hclog.Logger) *Currency {
	return &Currency{log: l, rates: r}
}

func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle GetRate", "base", rr.GetBase(), "destination", rr.GetDestination())

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())

	if err != nil {
		return nil, err
	}

	return &protos.RateResponse{Rate: rate}, nil
}
