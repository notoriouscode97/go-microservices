package handlers

import (
	"context"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/data"
	"net/http"
)

// MiddlewareValidateProduct validates the product in the request and calls next if ok
func (p *Products) MiddlewareValidateProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		prod := &data.Product{}

		err := data.FromJSON(prod, r.Body)
		if err != nil {
			p.l.Error("Deserializing product", "error", err)

			rw.WriteHeader(http.StatusBadRequest)
			_ = data.ToJSON(&GenericError{Message: err.Error()}, rw)
			return
		}

		// validate the product
		errs := p.v.Validate(prod)
		if len(errs) != 0 {
			p.l.Error("Validating product", "error", errs)

			// return the validation messages as an array
			rw.WriteHeader(http.StatusUnprocessableEntity)
			_ = data.ToJSON(&ValidationError{Messages: errs.Errors()}, rw)
			return
		}

		// add the product to the context
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		r = r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}

// MiddlewareValidateOrder validates the order in the request and calls the next handler if validation succeeds
func (p *Products) MiddlewareValidateOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")

		order := &data.OrderRequest{}

		// Deserialize request body into OrderRequest
		err := data.FromJSON(order, r.Body)
		if err != nil {
			p.l.Error("Deserializing order", "error", err)

			rw.WriteHeader(http.StatusBadRequest)
			_ = data.ToJSON(&GenericError{Message: err.Error()}, rw)
			return
		}

		// Validate the order
		errs := p.v.Validate(order)
		if len(errs) != 0 {
			p.l.Error("Validating order", "error", errs)

			// Return the validation messages as an array
			rw.WriteHeader(http.StatusUnprocessableEntity)
			_ = data.ToJSON(&ValidationError{Messages: errs.Errors()}, rw)
			return
		}

		// Add the validated order to the context
		ctx := context.WithValue(r.Context(), KeyOrder{}, order)
		r = r.WithContext(ctx)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(rw, r)
	})
}
