package handlers

import (
	"errors"
	"fmt"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/data"
	"net/http"
)

// swagger:route PUT /products products updateProduct
// Update a products details
//
// responses:
//	204: noContentResponse
//  404: errorResponse
//  422: errorValidation

// Update handles PUT requests to update products
func (p *Products) Update(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")

	// fetch the product from the context
	prod := r.Context().Value(KeyProduct{}).(*data.Product)

	v := data.NewValidation()

	// Validate the product
	if err := v.Validate(prod); err != nil { // Implement this validation function
		p.l.Error("Validation error:", err)
		rw.WriteHeader(http.StatusUnprocessableEntity) // 422
		errString := fmt.Sprintf("Validation error: %s", err.Errors())
		_ = data.ToJSON(&GenericError{Message: errString}, rw)
		return
	}
	p.l.Debug("Updating record id", prod.ID)

	err := p.productDB.UpdateProduct(prod)
	if errors.Is(err, data.ErrProductNotFound) {
		p.l.Error("Product not found", err)

		rw.WriteHeader(http.StatusNotFound)
		_ = data.ToJSON(&GenericError{Message: "Product not found in database"}, rw)
		return
	}

	if err != nil {
		p.l.Error("Error updating product", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}

	// write the no content success header
	rw.WriteHeader(http.StatusNoContent)
}
