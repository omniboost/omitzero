# omitzero

Go's `encoding/json` supports `omitzero` — a struct tag option that omits a field when it implements `IsZero() bool` and that method returns `true`. `encoding/xml` has no such option.

This package fills that gap. It provides:

- **`MarshalXML`** — a drop-in replacement for the `xml.Marshaler` interface body that adds `omitzero` support for XML.
- **`MarshalJSON`** — a thin wrapper around `encoding/json` for API symmetry. It does not add any behaviour; Go 1.24+ handles JSON `omitzero` natively.

## Installation

```sh
go get github.com/omniboost/omitzero
```

## Usage

### 1. Implement `IsZeroer` on your type

```go
type Amount struct {
    Value    float64
    Currency string
}

func (a Amount) IsZero() bool {
    return a.Value == 0 && a.Currency == ""
}
```

### 2. Tag struct fields with `omitzero`

```go
type Invoice struct {
    XMLName xml.Name `xml:"Invoice"`
    Number  string   `xml:"Number"`
    Total   Amount   `xml:"Total,omitzero"`
    Tax     Amount   `xml:"Tax,omitzero"` // omitted when zero
}
```

### 3. Wire in `MarshalXML`

Implement `xml.Marshaler` on your struct and delegate to this package:

```go
func (inv Invoice) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    return omitzero.MarshalXML(inv, e, start)
}
```

That is all. `xml.Marshal` and `xml.Encoder` will now honour `omitzero` for any field whose type returns `true` from `IsZero()`.

## Examples

### XML — zero fields are omitted

```go
inv := Invoice{
    Number: "INV-001",
    Total:  Amount{Value: 100, Currency: "EUR"},
    // Tax is zero — will be omitted
}

out, _ := xml.MarshalIndent(inv, "", "  ")
fmt.Println(string(out))
```

Output:

```xml
<Invoice>
  <Number>INV-001</Number>
  <Total>
    <Value>100</Value>
    <Currency>EUR</Currency>
  </Total>
</Invoice>
```

`Tax` is absent because `Amount{}.IsZero()` returned `true`.

> See [`ExampleMarshalXML`](example_test.go) for the runnable version.

### JSON — native passthrough

For JSON, tag fields with `json:",omitzero"` as usual. `encoding/json` (Go 1.24+) handles this natively via `IsZero()`. `MarshalJSON` delegates directly to `json.Marshal` without additional logic.

```go
type Product struct {
    Name     string `json:"name"`
    Price    Amount `json:"price,omitzero"`
    Discount Amount `json:"discount,omitzero"` // omitted when zero
}

p := Product{
    Name:  "Widget",
    Price: Amount{Value: 9.99, Currency: "EUR"},
}

out, _ := omitzero.MarshalJSON(p)
fmt.Println(string(out))
// {"name":"Widget","price":{"value":9.99,"currency":"EUR"}}
```

> See [`ExampleMarshalJSON`](example_test.go) for the runnable version.

## How it works

`encoding/xml` does not call `IsZero()` and does not recognise `omitzero` as a tag option. The workaround operates before the standard encoder is invoked:

1. Reflect over the struct fields.
2. For any field tagged `xml:",omitzero"`, check whether it is nil or whether it implements `IsZeroer` and `IsZero()` returns `true`.
3. If so, swap the field's tag to `xml:"-"`, which tells the standard encoder to skip it.
4. Pass the modified struct to `e.EncodeElement`.

The original struct is never mutated; the tag swap happens on a copy of the `reflect.StructField` slice.
