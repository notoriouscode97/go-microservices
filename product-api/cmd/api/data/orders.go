package data

type OrderRequest struct {
	Email string      `json:"email"`
	Items []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
