// Package jsonmodel mirrors pb.Order field-for-field so the JSON and
// Protobuf benchmarks are encoding the exact same information.
package jsonmodel

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type Item struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID              string            `json:"id"`
	CustomerName    string            `json:"customer_name"`
	CustomerEmail   string            `json:"customer_email"`
	CreatedAt       int64             `json:"created_at"`
	ShippingAddress Address           `json:"shipping_address"`
	Items           []Item            `json:"items"`
	Total           float64           `json:"total"`
	Status          string            `json:"status"`
	Metadata        map[string]string `json:"metadata"`
}
