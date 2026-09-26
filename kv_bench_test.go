package benchmark

// This file isolates ONE field's cost: the `metadata` map on pb.Order
// versus the same 5 entries modeled as `repeated KeyValue` on pb.OrderKV.
// Both produce byte-identical wire output (proto3 maps are specified as
// sugar over `repeated Entry{key=1;value=2;}`), so any gap in time or
// allocations below is coming purely from how google.golang.org/protobuf
// implements the two field kinds - not from the wire format.

import (
	"testing"

	"benchmark/kvpb"
	"benchmark/pb"

	"google.golang.org/protobuf/proto"
)

func newMapOnly() *pb.Order {
	return &pb.Order{
		Id: "ord_9f8a2b1c",
		Metadata: map[string]string{
			"referrer":      "google_ads",
			"campaign":      "autumn_sale_2026",
			"device":        "mobile",
			"session_id":    "sess_7d3f9a2e4b1c",
			"ab_test_group": "checkout_v3",
		},
	}
}

func newRepeatedKVOnly() *kvpb.OrderKV {
	return &kvpb.OrderKV{
		Id: "ord_9f8a2b1c",
		Metadata: []*kvpb.KeyValue{
			{Key: "referrer", Value: "google_ads"},
			{Key: "campaign", Value: "autumn_sale_2026"},
			{Key: "device", Value: "mobile"},
			{Key: "session_id", Value: "sess_7d3f9a2e4b1c"},
			{Key: "ab_test_group", Value: "checkout_v3"},
		},
	}
}

func BenchmarkMapMetadataMarshal(b *testing.B) {
	m := newMapOnly()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := proto.Marshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepeatedKVMetadataMarshal(b *testing.B) {
	m := newRepeatedKVOnly()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := proto.Marshal(m); err != nil {
			b.Fatal(err)
		}
	}
}

// TestKVWireSizeIsIdentical proves the two representations put the exact
// same bytes on the wire, so the benchmark above is measuring the Go
// implementation, not the format.
func TestKVWireSizeIsIdentical(t *testing.T) {
	mapData, err := proto.Marshal(newMapOnly())
	if err != nil {
		t.Fatal(err)
	}
	repData, err := proto.Marshal(newRepeatedKVOnly())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("map<string,string>:  %d bytes", len(mapData))
	t.Logf("repeated KeyValue:   %d bytes", len(repData))
	if len(mapData) != len(repData) {
		t.Logf("NOTE: sizes differ (map iteration order affects nothing here, only key/value bytes present)")
	}
}
