package benchmark

import (
	"encoding/json"
	"testing"

	"benchmark/jsonmodel"
	"benchmark/pb"

	"google.golang.org/protobuf/proto"
)

// Both constructors build the *same* order, field for field, so the two
// formats are encoding identical information. This is the detail most
// "JSON vs Protobuf" benchmarks on the internet get wrong.

func newJSONOrder() *jsonmodel.Order {
	return &jsonmodel.Order{
		ID:            "ord_9f8a2b1c",
		CustomerName:  "Nirmal Das",
		CustomerEmail: "bronirmaldaezz@example.com",
		CreatedAt:     1758870000,
		ShippingAddress: jsonmodel.Address{
			Street:  "14/2 Park Street",
			City:    "Kolkata",
			Zip:     "700016",
			Country: "IN",
		},
		Items: []jsonmodel.Item{
			{SKU: "SKU-1001", Name: "Mechanical Keyboard", Quantity: 1, Price: 89.99},
			{SKU: "SKU-1002", Name: "USB-C Hub", Quantity: 2, Price: 24.50},
			{SKU: "SKU-1003", Name: "27in Monitor", Quantity: 1, Price: 249.00},
		},
		Total:  387.99,
		Status: "processing",
		Metadata: map[string]string{
			"referrer":      "google_ads",
			"campaign":      "autumn_sale_2026",
			"device":        "mobile",
			"session_id":    "sess_7d3f9a2e4b1c",
			"ab_test_group": "checkout_v3",
		},
	}
}

func newProtoOrder() *pb.Order {
	return &pb.Order{
		Id:            "ord_9f8a2b1c",
		CustomerName:  "Nirmal Das",
		CustomerEmail: "bronirmaldaezz@example.com",
		CreatedAt:     1758870000,
		ShippingAddress: &pb.Address{
			Street:  "14/2 Park Street",
			City:    "Kolkata",
			Zip:     "700016",
			Country: "IN",
		},
		Items: []*pb.Item{
			{Sku: "SKU-1001", Name: "Mechanical Keyboard", Quantity: 1, Price: 89.99},
			{Sku: "SKU-1002", Name: "USB-C Hub", Quantity: 2, Price: 24.50},
			{Sku: "SKU-1003", Name: "27in Monitor", Quantity: 1, Price: 249.00},
		},
		Total:  387.99,
		Status: "processing",
		Metadata: map[string]string{
			"referrer":      "google_ads",
			"campaign":      "autumn_sale_2026",
			"device":        "mobile",
			"session_id":    "sess_7d3f9a2e4b1c",
			"ab_test_group": "checkout_v3",
		},
	}
}

// ---------- ENCODING ----------

func BenchmarkJSONMarshal(b *testing.B) {
	o := newJSONOrder()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(o); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtoMarshal(b *testing.B) {
	o := newProtoOrder()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := proto.Marshal(o); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- DECODING ----------

func BenchmarkJSONUnmarshal(b *testing.B) {
	data, err := json.Marshal(newJSONOrder())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var o jsonmodel.Order
		if err := json.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtoUnmarshal(b *testing.B) {
	data, err := proto.Marshal(newProtoOrder())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var o pb.Order
		if err := proto.Unmarshal(data, &o); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------- WIRE SIZE ----------

func TestPayloadSize(t *testing.T) {
	jsonData, err := json.Marshal(newJSONOrder())
	if err != nil {
		t.Fatal(err)
	}
	protoData, err := proto.Marshal(newProtoOrder())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("JSON payload:     %d bytes", len(jsonData))
	t.Logf("Protobuf payload: %d bytes", len(protoData))
	t.Logf("Protobuf is %.1f%% smaller than JSON", 100*(1-float64(len(protoData))/float64(len(jsonData))))
}
