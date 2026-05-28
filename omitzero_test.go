package omitempty_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	omitzero "github.com/omniboost/omitzero"
)

// xmlOrder mixes fields with and without omitzero so we can test both paths.
type xmlOrder struct {
	XMLName  xml.Name `xml:"Order"`
	ID       string   `xml:"ID"`
	Total    Amount   `xml:"Total,omitzero"`    // omitted when zero
	Discount Amount   `xml:"Discount,omitzero"` // omitted when zero
	Fee      Amount   `xml:"Fee"`               // no omitzero — never omitted
}

func (o xmlOrder) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return omitzero.MarshalXML(o, e, start)
}

// xmlOrderPtr has a pointer field to test nil detection.
type xmlOrderPtr struct {
	XMLName xml.Name `xml:"Order"`
	ID      string   `xml:"ID"`
	Extra   *Amount  `xml:"Extra,omitzero"`
}

func (o xmlOrderPtr) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return omitzero.MarshalXML(o, e, start)
}

// jsonProduct mirrors xmlOrder for the JSON path.
type jsonProduct struct {
	Name     string `json:"name"`
	Price    Amount `json:"price,omitzero"`    // omitted when zero
	Discount Amount `json:"discount,omitzero"` // omitted when zero
	Fee      Amount `json:"fee"`               // no omitzero — never omitted
}

// --- XML tests ---

func TestMarshalXML_ZeroFieldWithOmitzeroIsOmitted(t *testing.T) {
	order := xmlOrder{
		ID:    "ORD-1",
		Total: Amount{Value: 100, Currency: "EUR"},
		// Discount is zero
	}
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "<Discount>") {
		t.Errorf("zero Discount with omitzero should be omitted; got: %s", out)
	}
}

func TestMarshalXML_NonZeroFieldWithOmitzeroIsPresent(t *testing.T) {
	order := xmlOrder{
		ID:       "ORD-1",
		Total:    Amount{Value: 100, Currency: "EUR"},
		Discount: Amount{Value: 10, Currency: "EUR"},
	}
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<Discount>") {
		t.Errorf("non-zero Discount with omitzero should be present; got: %s", out)
	}
}

// TestMarshalXML_ZeroFieldWithoutOmitzeroIsPresent verifies that a zero-value
// struct field is NOT omitted when the omitzero tag option is absent.
func TestMarshalXML_ZeroFieldWithoutOmitzeroIsPresent(t *testing.T) {
	order := xmlOrder{
		ID:    "ORD-1",
		Total: Amount{Value: 100, Currency: "EUR"},
		// Fee is zero and has no omitzero tag — must still appear
	}
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<Fee>") {
		t.Errorf("zero Fee without omitzero tag should still be present; got: %s", out)
	}
}

func TestMarshalXML_NilPointerWithOmitzeroIsOmitted(t *testing.T) {
	order := xmlOrderPtr{ID: "ORD-1", Extra: nil}
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "<Extra>") {
		t.Errorf("nil Extra with omitzero should be omitted; got: %s", out)
	}
}

func TestMarshalXML_NonNilPointerWithOmitzeroIsPresent(t *testing.T) {
	order := xmlOrderPtr{
		ID:    "ORD-1",
		Extra: &Amount{Value: 5, Currency: "EUR"},
	}
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<Extra>") {
		t.Errorf("non-nil Extra with omitzero should be present; got: %s", out)
	}
}

func TestMarshalXML_AllOmitzeroFieldsZero(t *testing.T) {
	order := xmlOrder{ID: "ORD-1"} // Total, Discount zero; Fee zero but untagged
	out, err := xml.Marshal(order)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "<Total>") {
		t.Errorf("zero Total with omitzero should be omitted; got: %s", s)
	}
	if strings.Contains(s, "<Discount>") {
		t.Errorf("zero Discount with omitzero should be omitted; got: %s", s)
	}
	if !strings.Contains(s, "<Fee>") {
		t.Errorf("zero Fee without omitzero should still be present; got: %s", s)
	}
}

func TestMarshalXML_NonStructPassthrough(t *testing.T) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	start := xml.StartElement{Name: xml.Name{Local: "Value"}}
	if err := omitzero.MarshalXML("hello", enc, start); err != nil {
		t.Fatal(err)
	}
	enc.Flush()
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("non-struct value should pass through unchanged; got: %s", buf.String())
	}
}

// --- JSON tests ---

func TestMarshalJSON_ZeroFieldWithOmitzeroIsOmitted(t *testing.T) {
	p := jsonProduct{
		Name:  "Widget",
		Price: Amount{Value: 9.99, Currency: "EUR"},
		// Discount is zero
	}
	out, err := omitzero.MarshalJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["discount"]; ok {
		t.Errorf("zero Discount with omitzero should be omitted; got: %s", out)
	}
}

func TestMarshalJSON_NonZeroFieldWithOmitzeroIsPresent(t *testing.T) {
	p := jsonProduct{
		Name:     "Widget",
		Price:    Amount{Value: 9.99, Currency: "EUR"},
		Discount: Amount{Value: 1.00, Currency: "EUR"},
	}
	out, err := omitzero.MarshalJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["discount"]; !ok {
		t.Errorf("non-zero Discount with omitzero should be present; got: %s", out)
	}
}

// TestMarshalJSON_ZeroFieldWithoutOmitzeroIsPresent verifies that a zero-value
// struct field is NOT omitted when the omitzero tag option is absent.
func TestMarshalJSON_ZeroFieldWithoutOmitzeroIsPresent(t *testing.T) {
	p := jsonProduct{
		Name:  "Widget",
		Price: Amount{Value: 9.99, Currency: "EUR"},
		// Fee is zero and has no omitzero tag — must still appear
	}
	out, err := omitzero.MarshalJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["fee"]; !ok {
		t.Errorf("zero Fee without omitzero tag should still be present; got: %s", out)
	}
}
