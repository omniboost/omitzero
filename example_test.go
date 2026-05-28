package omitempty_test

import (
	"encoding/xml"
	"fmt"

	omitzero "github.com/omniboost/omitzero"
)

// Amount holds a monetary value. It implements IsZeroer so that fields tagged
// with omitzero are omitted when both Value and Currency are unset.
type Amount struct {
	Value    float64 `xml:"Value" json:"value,omitzero"`
	Currency string  `xml:"Currency" json:"currency,omitzero"`
}

func (a Amount) IsZero() bool {
	return a.Value == 0 && a.Currency == ""
}

// Invoice uses omitzero on its XML fields. It implements xml.Marshaler so the
// encoder routes through omitzero.MarshalXML.
type Invoice struct {
	XMLName xml.Name `xml:"Invoice"`
	Number  string   `xml:"Number"`
	Total   Amount   `xml:"Total,omitzero"`
	Tax     Amount   `xml:"Tax,omitzero"`
}

func (inv Invoice) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return omitzero.MarshalXML(inv, e, start)
}

// ExampleMarshalXML shows that zero-value fields tagged xml:",omitzero" are
// omitted from output. Tax is zero and will not appear; Total is non-zero and
// will be included.
func ExampleMarshalXML() {
	inv := Invoice{
		Number: "INV-001",
		Total:  Amount{Value: 100, Currency: "EUR"},
		// Tax is zero — will be omitted
	}

	out, _ := xml.MarshalIndent(inv, "", "  ")
	fmt.Println(string(out))
	// Output:
	// <Invoice>
	//   <Number>INV-001</Number>
	//   <Total>
	//     <Value>100</Value>
	//     <Currency>EUR</Currency>
	//   </Total>
	// </Invoice>
}

// ExampleMarshalJSON shows that MarshalJSON delegates to encoding/json, which
// natively omits zero-value fields tagged json:",omitzero" via IsZero().
func ExampleMarshalJSON() {
	type Product struct {
		Name     string `json:"name"`
		Price    Amount `json:"price,omitzero"`
		Discount Amount `json:"discount,omitzero"`
	}

	p := Product{
		Name:  "Widget",
		Price: Amount{Value: 9.99, Currency: "EUR"},
		// Discount is zero — will be omitted by encoding/json natively
	}

	out, _ := omitzero.MarshalJSON(p)
	fmt.Println(string(out))
	// Output:
	// {"name":"Widget","price":{"value":9.99,"currency":"EUR"}}
}
