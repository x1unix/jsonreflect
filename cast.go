package jsonreflect

import "fmt"

func newInvalidValueError(gotType, wantType Type) error {
	return fmt.Errorf("cannot convert jsonreflect.Value of type %s to %s", gotType.String(), wantType.String())
}

// ToObject casts generic value to jsonreflect.Object.
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

// ToNumber casts generic value to jsonreflect.Number.
//
// Method only supports number and string values.
func ToNumber(v Value, bitSize int) (*Number, error) {
	switch t := TypeOf(v); t {
	case TypeNumber:
		return v.(*Number), nil
	case TypeString:
		strval, err := v.String()
		if err != nil {
			return nil, err
		}

		numval, err := numberValueFromString(v.Ref(), strval, bitSize)
		if err != nil {
			return nil, fmt.Errorf("cannot cast %s value %q to %s", t, strval, TypeNumber)
		}

		return numval, nil
	default:
		return nil, fmt.Errorf("cannot cast %s value to %s", t, TypeNumber)
	}
}

// ToArray casts generic value to jsonreflect.Array.
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
