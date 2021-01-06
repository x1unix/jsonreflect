package jsonreflect

import (
	"io"
	"sort"
)

// Object represents key-value pair of object field and value
type Object struct {
	baseValue

	// Items is key-value pair of object values
	Items map[string]Value
}

func newObject(start, end int, items map[string]Value) *Object {
	return &Object{
		baseValue: newBaseValue(start, end),
		Items:     items,
	}
}

// Type implements jsonreflect.Value
func (_ Object) Type() Type {
	return TypeObject
}

// Keys returns sorted list of object keys
func (o Object) Keys() []string {
	if len(o.Items) == 0 {
		return nil
	}

	keys := make([]string, 0, len(o.Items))
	for k := range o.Items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// HasKey checks if key exists in object
func (o Object) HasKey(keyName string) bool {
	_, ok := o.Items[keyName]
	return ok
}

func (o Object) marshal(w io.Writer, mf *marshalFormatter) error {
	if len(o.Items) == 0 {
		return mf.write(w, []byte{tokenObjectStart, tokenObjectClose})
	}

	err := mf.writeOpenClause(w, tokenObjectStart)
	if err != nil {
		return err
	}

	keys := o.Keys()
	childFmt := mf.childFormatter()
	lastIndex := len(keys) - 1
	for i, key := range keys {
		value := o.Items[key]
		err = childFmt.writePropertyName(w, key)
		if err != nil {
			return err
		}

		err = value.marshal(w, childFmt)
		if err != nil {
			return err
		}

		err = mf.writeElementDelimiter(w, i == lastIndex)
		if err != nil {
			return err
		}
	}

	return mf.write(w, []byte{tokenObjectClose})
}

// ToMap returns key-value pair of items as interface value
func (o Object) ToMap() map[string]interface{} {
	m := make(map[string]interface{}, len(o.Items))
	for k, v := range o.Items {
		m[k] = v.Interface()
	}
	return m
}

// Interface() implements json.Value
func (o Object) Interface() interface{} {
	return o.ToMap()
}
