# JSON vs Protobuf Serialization Benchmark in Go

A benchmark comparison in **JSON** and **Protocol Buffers
(Protobuf)** serialization in Go across:

-   Marshal / encoding performance
-   Unmarshal / decoding performance
-   Allocations
-   Bytes allocated per operation
-   Wire/payload size
-   CPU profiles using `pprof`
-   A Protobuf `map<string,string>` representation vs an equivalent
    repeated key/value representation

The benchmark intentionally uses the **same logical order data** for
JSON and Protobuf so that the comparison is about the serialization
format and its Go implementation rather than different payloads.

------------------------------------------------------------------------

## TL;DR

For the tested `Order` payload 


  Operation                    JSON           Protobuf           Observed
                                                               difference

  Marshal                2289 ns/op         2044 ns/op   Protobuf \~12.0%
                                                               lower time

  Unmarshal              9605 ns/op         2717 ns/op   Protobuf \~3.53×
                                                               lower time

  Marshal memory          1008 B/op           704 B/op   Protobuf \~30.2%
                                                              fewer bytes

  Marshal              12 allocs/op       21 allocs/op     JSON has fewer
  allocations                                                 allocations

  Unmarshal               1568 B/op          1592 B/op    Nearly the same
  memory                                               

  Unmarshal            43 allocs/op       54 allocs/op     JSON has fewer
  allocations                                                 allocations

  Payload size            616 bytes          372 bytes     Protobuf 39.6%
                                                                  smaller
  -----------------------------------------------------------------------

The important point is that **fewer allocations does not automatically
mean lower latency**. In this benchmark, Protobuf unmarshal performs
substantially faster despite reporting more allocations per operation.

------------------------------------------------------------------------

## 1. What is being benchmarked?

The benchmark uses an `Order` containing:

-   Order ID
-   Customer name and email
-   Timestamp
-   Shipping address
-   Three order items
-   Total price
-   Status
-   Five metadata entries

The JSON model and Protobuf model represent the same logical information
field-for-field.

The Protobuf schema contains:

``` proto
message Address {
  string street  = 1;
  string city    = 2;
  string zip     = 3;
  string country = 4;
}

message Item {
  string sku      = 1;
  string name     = 2;
  int32  quantity = 3;
  double price    = 4;
}

message Order {
  string  id               = 1;
  string  customer_name    = 2;
  string  customer_email   = 3;
  int64   created_at       = 4;
  Address shipping_address = 5;
  repeated Item items      = 6;
  double  total            = 7;
  string  status           = 8;
  map<string, string> metadata = 9;
}
```

The generated Go Protobuf type contains the corresponding `Address`,
`Item`, and `Order` structures, including the `map[string]string`
metadata field. The generated code was produced with
`protoc-gen-go v1.32.0` and `protoc v3.21.12`. [Generated Protobuf code
reference](./pb/order.pb.go)

------------------------------------------------------------------------

# 2. Benchmark setup

The benchmark creates equivalent JSON and Protobuf orders.

For JSON:

``` go
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
```

The Protobuf constructor contains the same values using the generated
`pb.Order`, `pb.Address`, and `pb.Item` types.

This matters because a benchmark comparing different structures would
not isolate serialization overhead.

------------------------------------------------------------------------

# 3. Benchmark methodology

The benchmarks use Go's standard `testing` package and call:

``` go
b.ReportAllocs()
```

This reports:

-   `ns/op` --- nanoseconds required per operation
-   `B/op` --- bytes allocated per operation
-   `allocs/op` --- number of allocations per operation

The benchmark functions repeatedly perform only the operation being
measured.

### Marshal

``` go
for i := 0; i < b.N; i++ {
    _, err := json.Marshal(o)
}
```

and:

``` go
for i := 0; i < b.N; i++ {
    _, err := proto.Marshal(o)
}
```

### Unmarshal

The encoded payload is generated **before** the timed loop.

That is important: the benchmark measures decoding, not encoding +
decoding together.

``` go
data, err := json.Marshal(newJSONOrder())

b.ReportAllocs()

for i := 0; i < b.N; i++ {
    var o jsonmodel.Order
    json.Unmarshal(data, &o)
}
```

The Protobuf benchmark follows the same structure.

------------------------------------------------------------------------

# 4. Payload size

The first result measures the serialized wire payload.

### Result

``` text
JSON payload:     616 bytes
Protobuf payload: 372 bytes
Protobuf is 39.6% smaller than JSON
```

### Screenshot

```{=html}
<!-- SCREENSHOT 1: Paste the payload-size + wire-size-identical benchmark screenshot here -->
```
![Payload size benchmark](PASTE_SCREENSHOT_1_LINK_HERE)

### Interpretation

For this particular payload:

``` text
JSON       = 616 bytes
Protobuf   = 372 bytes
Difference = 244 bytes
```

So Protobuf uses approximately **39.6% less serialized payload space**.

This is a wire-format result for this specific schema and payload. It
should not be interpreted as a universal compression ratio for every
JSON/Protobuf workload.

------------------------------------------------------------------------

# 5. Marshal benchmark

Marshal means:

``` text
Go object → serialized bytes
```

The benchmark was run with:

``` bash
go test -run=^$ -bench=BenchmarkJSONMarshal
```

and:

``` bash
go test -run=^$ -bench=BenchmarkProtoMarshal
```

## Results

### JSON

``` text
BenchmarkJSONMarshal-8
498070
2289 ns/op
1008 B/op
12 allocs/op
```

### Protobuf

``` text
BenchmarkProtoMarshal-8
549082
2044 ns/op
704 B/op
21 allocs/op
```

### Comparison

  Metric                      JSON       Protobuf
  ----------------- -------------- --------------
  Time                  2289 ns/op     2044 ns/op
  Bytes allocated        1008 B/op       704 B/op
  Allocations         12 allocs/op   21 allocs/op

Relative to JSON:

-   Protobuf marshal time is about **12.0% lower**.
-   Protobuf allocates about **30.2% fewer bytes**.
-   Protobuf performs **more individual allocations** in this benchmark:
    21 vs 12.

That last point is worth highlighting.

> Allocation count and total allocated bytes are different measurements.

A program can perform more allocation events while allocating fewer
total bytes.

### Screenshot

```{=html}
<!-- SCREENSHOT 2: Paste the marshal benchmark result screenshot here -->
```
![Marshal benchmark](PASTE_SCREENSHOT_2_LINK_HERE)

------------------------------------------------------------------------

# 6. Unmarshal benchmark

Unmarshal means:

``` text
serialized bytes → Go object
```

This is where the largest latency difference appears in this benchmark.

## JSON result

``` text
BenchmarkJSONUnmarshal-8
117398
9605 ns/op
1568 B/op
43 allocs/op
```

## Protobuf result

``` text
BenchmarkProtoUnmarshal-8
418088
2717 ns/op
1592 B/op
54 allocs/op
```

### Comparison

  Metric                      JSON       Protobuf
  ----------------- -------------- --------------
  Time                  9605 ns/op     2717 ns/op
  Bytes allocated        1568 B/op      1592 B/op
  Allocations         43 allocs/op   54 allocs/op

### Time difference

``` text
9605 / 2717 ≈ 3.53
```

So, for this benchmark:

**Protobuf unmarshaling is about 3.53× faster than JSON unmarshaling.**

The allocation numbers tell a more interesting story:

``` text
JSON       43 allocs/op
Protobuf   54 allocs/op
```

Yet the Protobuf decoder is substantially faster.

This demonstrates why looking only at `allocs/op` can lead to incomplete
conclusions about performance.

### Screenshot --- JSON unmarshal

```{=html}
<!-- SCREENSHOT 3: Paste the JSON unmarshal benchmark screenshot here -->
```
![JSON unmarshal benchmark](PASTE_SCREENSHOT_3_LINK_HERE)

### Screenshot --- Protobuf unmarshal

```{=html}
<!-- SCREENSHOT 4: Paste the Protobuf unmarshal benchmark screenshot here -->
```
![Protobuf unmarshal benchmark](PASTE_SCREENSHOT_4_LINK_HERE)

------------------------------------------------------------------------

# 7. Why is Protobuf unmarshal faster here?

The benchmark result shows a large decoding difference, but the
benchmark itself does not prove a single cause.

The CPU profile gives us more detail about where execution time was
sampled.

For JSON, the profile includes functions such as:

``` text
encoding/json.stateInString
encoding/json.(*decodeState).object
encoding/json.unquoteBytes
encoding/json.(*decodeState).rescanLiteral
encoding/json.checkValid
encoding/json.stateEndValue
```

For Protobuf, the profile includes:

``` text
google.golang.org/protobuf/internal/impl.(*MessageInfo).unmarshalPointer
google.golang.org/protobuf/internal/impl.consumeMap
runtime.mallocgc
```

A key difference is that JSON has to interpret a text representation
containing field names, strings, punctuation, literals, and JSON syntax.

Protobuf instead uses a binary wire representation where field numbers
and wire types guide decoding.

The profile is evidence about where CPU samples landed; it should not be
interpreted as saying every sample in `runtime.*` is caused exclusively
by the serialization library.

------------------------------------------------------------------------

# 8. CPU profiling --- JSON Unmarshal

The JSON benchmark was run with:

``` bash
go test -run=^$ \
  -bench=BenchmarkJSONUnmarshal \
  -cpuprofile=json.prof
```

Then:

``` bash
go tool pprof -top -nodecount=10 json.prof
```

### Important profile results

``` text
encoding/json.stateInString       70ms   6.73%
encoding/json.(*decodeState).object
                                   60ms   flat
                                   600ms cumulative
encoding/json.unquoteBytes         60ms   5.77%
encoding/json.(*decodeState).rescanLiteral
                                   50ms   4.81%
encoding/json.checkValid            40ms   3.85% flat
                                   180ms cumulative
encoding/json.stateEndValue         40ms   3.85%
```

The profile reported:

``` text
Duration: 1.42s
Total samples: 1040ms
```

### Screenshot

```{=html}
<!-- SCREENSHOT 5: Paste the JSON CPU profile screenshot here -->
```
![JSON CPU profile](PASTE_SCREENSHOT_5_LINK_HERE)

### Reading `flat` vs `cum`

For example:

``` text
encoding/json.(*decodeState).object
flat = 60ms
cum  = 600ms
```

`flat` means the samples attributed directly to that function.

`cum` includes the function plus functions it called beneath it in the
call tree.

So a high cumulative value does **not** mean the function itself
directly consumed all that CPU time.

------------------------------------------------------------------------

# 9. CPU profiling --- Protobuf Unmarshal

The Protobuf benchmark was profiled with:

``` bash
go test -run=^$ \
  -bench=BenchmarkProtoUnmarshal \
  -cpuprofile=proto.prof
```

Then:

``` bash
go tool pprof -top -nodecount=10 proto.prof
```

### Important profile results

``` text
runtime.kevent
    450ms   37.82%

runtime.pthread_cond_wait
    190ms   15.97%

google.golang.org/protobuf/internal/impl.
(*MessageInfo).unmarshalPointer
     50ms flat
    390ms cumulative

google.golang.org/protobuf/internal/impl.consumeMap
     30ms flat
    190ms cumulative

runtime.mallocgc
     30ms   2.52%
```

The profile reported:

``` text
Duration: 1.31s
Total samples: 1190ms
```

### Screenshot

```{=html}
<!-- SCREENSHOT 6: Paste the Protobuf CPU profile screenshot here -->
```
![Protobuf CPU profile](PASTE_SCREENSHOT_6_LINK_HERE)

### Important observation

The profile contains substantial `runtime.*` activity.

That is normal in a real benchmark process, and those samples should not
simply be labeled as "Protobuf overhead."

For example:

``` text
runtime.kevent
runtime.pthread_cond_wait
runtime.pthread_cond_signal
runtime.pthread_cond_timedwait_relative_np
```

are runtime/OS scheduling and synchronization related activity.

The more directly relevant Protobuf functions in this profile are:

``` text
(*MessageInfo).unmarshalPointer
consumeMap
```

------------------------------------------------------------------------

# 10. Memory allocation profiles

CPU time tells us **where the processor spent sampled time**.

Memory profiling gives us another perspective:

``` text
Where are allocations coming from?
```

The benchmark was run with:

``` bash
go test -run=^$ \
  -bench=BenchmarkProtoMarshal \
  -memprofile=mem.prof
```

and:

``` bash
go test -run=^$ \
  -bench=BenchmarkJSONMarshal \
  -memprofile=mem.prof
```

The benchmark results themselves reported:

### Protobuf Marshal

``` text
2044 ns/op
704 B/op
21 allocs/op
```

### JSON Marshal

``` text
2289 ns/op
1008 B/op
12 allocs/op
```

### Screenshot --- Protobuf memory benchmark

```{=html}
<!-- SCREENSHOT 7: Paste the Protobuf memory benchmark screenshot here -->
```
![Protobuf memory benchmark](PASTE_SCREENSHOT_7_LINK_HERE)

### Screenshot --- JSON memory benchmark

```{=html}
<!-- SCREENSHOT 8: Paste the JSON memory benchmark screenshot here -->
```
![JSON memory benchmark](PASTE_SCREENSHOT_8_LINK_HERE)

------------------------------------------------------------------------

# 11. A subtle result: fewer allocations ≠ faster code

One of the most interesting results is:

  Metric        JSON Marshal   Protobuf Marshal
  ----------- -------------- ------------------
  ns/op                 2289           **2044**
  B/op                  1008            **704**
  allocs/op           **12**                 21

JSON performs fewer allocations:

``` text
12 vs 21
```

But Protobuf:

-   takes less time,
-   allocates fewer total bytes,
-   produces a smaller wire payload.

This is why performance analysis should not reduce everything to:

> "Fewer allocations = faster."

The cost of an allocation depends on its size, lifetime, object
structure, allocator behavior, GC interaction, and the work required
around it.

------------------------------------------------------------------------

# 12. Metadata map experiment

The project also contains a focused experiment around:

``` proto
map<string, string> metadata
```

The experiment compares:

``` proto
map<string, string>
```

with an explicit:

``` proto
repeated KeyValue metadata
```

where:

``` proto
message KeyValue {
  string key   = 1;
  string value = 2;
}
```

The goal is to isolate the implementation cost of the Go Protobuf map
representation from the underlying wire representation.

The test output showed:

``` text
map<string,string>: 149 bytes
repeated KeyValue:  149 bytes
```

and the test passed.

### Important detail

A Protobuf map field is represented on the wire using map-entry
messages. Therefore, the experiment is useful for separating:

``` text
wire representation
```

from:

``` text
Go runtime representation / decoding implementation
```

However, the current screenshot set does **not** include the dedicated
marshal benchmark numbers for:

``` text
BenchmarkMapMetadataMarshal
BenchmarkRepeatedKVMetadataMarshal
```

so this README does not invent or report those timings.

------------------------------------------------------------------------

# 13. Benchmark commands

## Run all tests

``` bash
go test -run=Test -v ./...
```

## Run JSON marshal benchmark

``` bash
go test -run=^$ -bench=BenchmarkJSONMarshal
```

## Run Protobuf marshal benchmark

``` bash
go test -run=^$ -bench=BenchmarkProtoMarshal
```

## Run JSON unmarshal benchmark with CPU profiling

``` bash
go test -run=^$ \
  -bench=BenchmarkJSONUnmarshal \
  -cpuprofile=json.prof
```

## Run Protobuf unmarshal benchmark with CPU profiling

``` bash
go test -run=^$ \
  -bench=BenchmarkProtoUnmarshal \
  -cpuprofile=proto.prof
```

## Inspect JSON CPU profile

``` bash
go tool pprof -top -nodecount=10 json.prof
```

## Inspect Protobuf CPU profile

``` bash
go tool pprof -top -nodecount=10 proto.prof
```

## Run Protobuf marshal with memory profiling

``` bash
go test -run=^$ \
  -bench=BenchmarkProtoMarshal \
  -memprofile=mem.prof
```

## Run JSON marshal with memory profiling

``` bash
go test -run=^$ \
  -bench=BenchmarkJSONMarshal \
  -memprofile=mem.prof
```

------------------------------------------------------------------------

# 14. Benchmark environment

The screenshots show:

``` text
OS:        darwin
Architecture: arm64
CPU:       Apple M2
Package:   benchmark
```

The benchmark was run on:

``` text
Apple M2
```

The exact Go version is not captured in the provided benchmark output,
so it is intentionally not listed here.

For reproducible comparisons, always record:

-   Go version
-   CPU
-   OS
-   architecture
-   protobuf library version
-   benchmark payload
-   benchmark duration / count
-   compiler flags if customized

------------------------------------------------------------------------

# 15. Results at a glance

``` text
                 JSON          Protobuf
------------------------------------------------
Payload          616 B         372 B
Marshal          2289 ns/op    2044 ns/op
Marshal B/op     1008 B        704 B
Marshal allocs   12            21

Unmarshal        9605 ns/op    2717 ns/op
Unmarshal B/op   1568 B        1592 B
Unmarshal allocs 43            54
```

The biggest measured difference is decoding:

``` text
JSON       9605 ns/op
Protobuf   2717 ns/op
```

which is approximately:

``` text
3.53×
```

for this particular benchmark.

------------------------------------------------------------------------

# 16. What this benchmark demonstrates

### 1. Payload size matters

Protobuf encoded this particular order in:

``` text
372 bytes
```

versus:

``` text
616 bytes
```

for JSON.

### 2. Encoding performance can be close

Marshal performance was:

``` text
JSON       2289 ns/op
Protobuf   2044 ns/op
```

The difference is much smaller than the unmarshal difference.

### 3. Decoding shows a much larger gap

The measured unmarshal latency was:

``` text
JSON       9605 ns/op
Protobuf   2717 ns/op
```

### 4. Allocation count alone is not enough

Protobuf had more allocations per operation in these tests while still
showing lower marshal latency and substantially lower unmarshal latency.

### 5. Profiling is necessary

The benchmark numbers tell us **what happened**.

`pprof` helps investigate **where the time was spent**.

For JSON, the profile shows significant activity in the JSON
parser/state-machine functions.

For Protobuf, the profile shows activity in generated/runtime
unmarshaling and map consumption, alongside Go runtime
scheduling/allocation functions.

------------------------------------------------------------------------

# 17. What this benchmark does NOT prove

This benchmark is intentionally narrow.

It does **not** prove that:

-   Protobuf is always faster than JSON.
-   Protobuf always allocates less.
-   Protobuf is always the right choice.
-   JSON is always slower in every workload.
-   The measured percentages will be identical on another CPU.
-   The measured numbers represent network-level end-to-end performance.
-   The profile percentages represent only serialization-library work.

Real systems can have very different results depending on:

-   payload size
-   nesting depth
-   number of repeated fields
-   number of map entries
-   string lengths
-   numeric fields
-   optional fields
-   compression
-   CPU architecture
-   Go version
-   protobuf implementation/version
-   GC behavior
-   network stack
-   application workload

------------------------------------------------------------------------

# 18. Key takeaway

The useful lesson isn't simply:

> "Protobuf is faster than JSON."

The more useful lesson is:

> **Measure the workload you actually care about.**

In this benchmark, Protobuf produced a smaller wire payload and showed
lower marshal latency and substantially lower unmarshal latency.

But the allocation profile also shows why performance analysis needs
multiple measurements:

``` text
latency
+ bytes allocated
+ allocation count
+ CPU profile
+ memory profile
+ payload size
```

Looking at only one metric can hide important behavior.

------------------------------------------------------------------------

## Project structure

A simplified structure of the benchmark:

``` text
serialize/
├── benchmark/
│   ├── benchmark_test.go
│   ├── kv_bench_test.go
│   ├── jsonmodel/
│   │   └── ...
│   ├── pb/
│   │   └── order.pb.go
│   └── kvpb/
│       └── ...
├── json.prof
├── proto.prof
└── mem.prof
```

The generated Protobuf code is produced from the `.proto` schema and
contains the generated `Address`, `Item`, and `Order` message
implementations. The generated `Order` includes the `metadata` map
field. [Generated code](./pb/order.pb.go)

------------------------------------------------------------------------

## Notes

These results are from a single benchmark environment and the
screenshots provided with this experiment. For a stronger performance
study, the next step would be to run repeated benchmarks with controlled
payload sizes and compare distributions using tools such as `benchstat`,
rather than relying on one run.
