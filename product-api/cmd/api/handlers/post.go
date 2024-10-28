package handlers

import (
	"fmt"
	"github.com/notoriouscode97/go-microservices/product-api/cmd/api/data"
	"net/http"
)

// swagger:route POST /products products createProduct
// Create a new product
//
// responses:
//	201: productResponse
//  422: errorValidation
//  500: errorResponse

// Create handles POST requests to add new products
func (p *Products) Create(rw http.ResponseWriter, r *http.Request) {
	// Fetch the product from the context
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

	p.l.Debug("Inserting product: %#v\n", prod)
	err := p.productDB.AddProduct(prod)

	if err != nil {
		p.l.Error("Product not created", err)
		rw.WriteHeader(http.StatusInternalServerError) // 500
		_ = data.ToJSON(&GenericError{Message: "error adding product"}, rw)
		return
	}

	// Write the status created success header
	rw.WriteHeader(http.StatusCreated) // 201

	// Include the created product in the response
	rw.Header().Set("Location", fmt.Sprintf("/products/%d", prod.ID)) // Set Location header
	if err := data.ToJSON(prod, rw); err != nil {
		p.l.Error("Error encoding response:", err)
		rw.WriteHeader(http.StatusInternalServerError) // 500
		_ = data.ToJSON(&GenericError{Message: "error encoding response"}, rw)
		return
	}
}
