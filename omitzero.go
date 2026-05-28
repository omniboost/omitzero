package omitempty

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strings"
)

// IsZeroer matches the interface used by encoding/json for the omitzero option.
type IsZeroer interface {
	IsZero() bool
}

// MarshalJSON is a no-op helper for symmetry with MarshalXML. encoding/json
// natively supports omitzero via IsZero() since Go 1.24, so this just
// delegates to json.Marshal without any special handling.
func MarshalJSON(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// MarshalXML is a workaround for encoding/xml lacking omitzero support.
// Fields tagged with xml:",omitzero" are omitted when they implement IsZeroer
// and IsZero() returns true.
func MarshalXML(obj interface{}, e *xml.Encoder, start xml.StartElement) error {
	st := reflect.TypeOf(obj)
	if st.Kind() != reflect.Struct {
		return e.EncodeElement(obj, start)
	}

	fs := []reflect.StructField{}
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		if len(f.PkgPath) != 0 { // skip unexported fields
			continue
		}
		fs = append(fs, f)
	}

	for i := range fs {
		if !fieldHasOption(fs[i], "xml", "omitzero") {
			continue
		}

		val := reflect.ValueOf(obj)
		f := val.Field(fs[i].Index[0]).Interface()

		if isNil(f) {
			fs[i].Tag = reflect.StructTag(`xml:"-"`)
			continue
		}

		if z, ok := f.(IsZeroer); ok && z.IsZero() {
			fs[i].Tag = reflect.StructTag(`xml:"-"`)
		}
	}

	st2 := reflect.StructOf(fs)
	v2 := reflect.ValueOf(obj).Convert(st2)
	return e.EncodeElement(v2.Interface(), start)
}

func fieldHasOption(field reflect.StructField, encoder, option string) bool {
	for _, opt := range strings.Split(field.Tag.Get(encoder), ",") {
		if opt == option {
			return true
		}
	}
	return false
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}
	return reflect.DeepEqual(a, reflect.Zero(reflect.TypeOf(a)).Interface())
}
