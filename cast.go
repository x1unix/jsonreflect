package jsonx

import "fmt"

func newInvalidValueError(gotType, wantType Type) error {
	return fmt.Errorf("cannot convert jsonx.Value of type %s to %s", gotType.String(), wantType.String())
}

// ToObject casts generic value to jsonx.Object.
// Passed value should be object type.
//
// Basically, it's alias to:
//
//	val, ok := v.(*Object)
func ToObject(v Value) (*Object, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot cast nil value to %s", TypeObject.String())
	}

	val, ok := v.(*Object)
	if !ok {
		return nil, newInvalidValueError(v.Type(), TypeObject)
	}
	return val, nil
}

// ToArray casts generic value to jsonx.Array.
// Passed value should be object type.
//
// Basically, it's alias to:
//
//	val, ok := v.(*Array)
func ToArray(v Value) (*Array, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot cast nil value to %s", TypeArray.String())
	}

	val, ok := v.(*Array)
	if !ok {
		return nil, newInvalidValueError(v.Type(), TypeArray)
	}
	return val, nil
}

// NewArray creates a new array of values
func NewArray(items ...Value) *Array {
	return &Array{
		Length: len(items),
		Items:  items,
	}
}
