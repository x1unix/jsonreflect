package jsonx

import (
	"fmt"
	"io"
	"strconv"
)

// Type represents value type
type Type uint

const (
	// TypeUnknown is invalid value type
	TypeUnknown Type = iota

	// TypeNull is null value type
	TypeNull

	// TypeBoolean is boolean value type
	TypeBoolean

	// TypeNumber is number value type
	TypeNumber

	// TypeString is string value type
	TypeString

	// TypeObject is object value type
	TypeObject

	// TypeArray is array value type
	TypeArray
)

// String returns value type as string
func (t Type) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	default:
		return "undefined"
	}
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

type Position struct {
	Start int
	End   int
}

func newPosition(start, end int) Position {
	return Position{Start: start, End: end}
}

type baseValue struct {
	// Position is value declaration position
	Position Position
}

func newBaseValue(start, end int) baseValue {
	return baseValue{newPosition(start, end)}
}

// Type implements json.Value
func (v baseValue) Type() Type {
	return TypeUnknown
}

// Ref implements json.Value
func (v baseValue) Ref() Position {
	return v.Position
}

// Value is abstract JSON document value
type Value interface {
	// Ref returns reference to value in source
	Ref() Position

	// Type returns value type
	Type() Type

	// Interface returns interface{} value
	Interface() interface{}

	// marshal serializes value with specified params
	marshal(io.Writer, *marshalFormatter) error
}

// String represents JSON string
type String struct {
	baseValue
	rawValue []byte
}

func newString(pos Position, val []byte) *String {
	return &String{
		baseValue: baseValue{pos},
		rawValue:  val,
	}
}

func (s String) marshal(w io.Writer, _ *marshalFormatter) error {
	_, err := w.Write(s.rawValue)
	return err
}

// Type implements jsonx.Value
func (_ String) Type() Type {
	return TypeString
}

// RawString returns quoted raw string
func (s String) RawString() string {
	return string(s.rawValue)
}

// String() returns unquoted string value
func (s String) String() (string, error) {
	str := s.RawString()
	v, err := strconv.Unquote(str)
	if err != nil {
		return "", fmt.Errorf("jsonx.String: failed to unquote raw string value '%s': %w", s.rawValue, err)
	}

	return v, nil
}

// Number returns number quoted in string
func (s String) Number() (*Number, error) {
	v, err := s.String()
	if err != nil {
		return nil, err
	}
	return ParseNumber(s.Position, v, 64)
}

// Interface() implements json.Value
func (s String) Interface() interface{} {
	v, err := s.String()
	if err != nil {
		return s.RawString()
	}
	return v
}

// Boolean is boolean value
type Boolean struct {
	baseValue
	Value bool
}

func newBoolean(pos Position, val bool) Boolean {
	return Boolean{
		baseValue: baseValue{
			pos,
		},
		Value: val,
	}
}

func (b Boolean) marshal(w io.Writer, _ *marshalFormatter) error {
	_, err := w.Write([]byte(strconv.FormatBool(b.Value)))
	return err
}

// Interface() implements json.Value
func (b Boolean) Interface() interface{} {
	return b.Value
}

// Type implements jsonx.Value
func (_ Boolean) Type() Type {
	return TypeBoolean
}

// Null is JSON null value
type Null struct {
	baseValue
}

// Type implements jsonx.Value
func (_ Null) Type() Type {
	return TypeNull
}

func (_ Null) marshal(w io.Writer, _ *marshalFormatter) error {
	_, err := w.Write([]byte("null"))
	return err
}

func newNull(pos Position) Null {
	return Null{baseValue{pos}}
}

// Interface() implements json.Value
func (n Null) Interface() interface{} {
	return nil
}
