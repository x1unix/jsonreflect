package jsonreflect

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// Accepts additional options to customise unmarshal process.
//
// Method supports the same tag and behavior as standard json.Unmarshal method.
//
// See UnmarshalValue documentation for information about extended behavior.
func Unmarshal(src []byte, dst interface{}, opts ...UnmarshalOption) error {
	value, err := ValueOf(src)
	if err != nil {
		return err
	}

	if TypeOf(value) == TypeNull {
		return nil
	}

	return UnmarshalValue(value, dst, opts...)
}

// ValueOf parses the JSON-encoded data and returns a document structure.
//
// Alias to NewParser().Parse()
func ValueOf(src []byte) (Value, error) {
	return NewParser(src).Parse()
}

// TypeOf returns value type.
//
// Returns TypeNull if nil value passed.
func TypeOf(v Value) Type {
	if v == nil {
		return TypeNull
	}
	return v.Type()
}
