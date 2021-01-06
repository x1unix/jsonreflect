package jsonx

import (
	"fmt"
	"math"
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

// Interface() implements json.Value
func (b Boolean) Interface() interface{} {
	return b.Value
}

// Type implements jsonx.Value
func (_ Boolean) Type() Type {
	return TypeBoolean
}

// Number represents json float64 number value
type Number struct {
	baseValue
	mantissa int64
	exponent uint64
	expoLen  int

	// IsFloat is floating point number flag
	IsFloat bool

	// IsSigned is signed number flag
	IsSigned bool
}

// Type implements jsonx.Value
func (_ Number) Type() Type {
	return TypeNumber
}

// Interface() implements json.Value
func (n Number) Interface() interface{} {
	if n.IsFloat {
		return n.Float64()
	}
	return n.Int()
}

// Float64 returns value as float64 number
func (n Number) Float64() float64 {
	if n.exponent == 0 {
		return float64(n.mantissa)
	}

	exponent := float64(n.exponent) / math.Pow10(n.expoLen)
	if n.mantissa < 0 {
		exponent *= -1
	}
	return float64(n.mantissa) + exponent
}

// Float32 returns value as float32 number
func (n Number) Float32() float32 {
	return float32(n.Float64())
}

// Int returns value as integer number
func (n Number) Int() int {
	return int(n.mantissa)
}

// Int64 returns value as int64 number
func (n Number) Int64() int64 {
	return n.mantissa
}

// Int32 returns value as int32 number
func (n Number) Int32() int32 {
	return int32(n.mantissa)
}

// Uint returns value as unsigned integer number
func (n Number) Uint() uint {
	return uint(n.mantissa)
}

// Uint32 returns value as uint32 number
func (n Number) Uint32() uint32 {
	return uint32(n.mantissa)
}

// Uint64 returns value as uint64 number
func (n Number) Uint64() uint64 {
	return uint64(n.mantissa)
}

// Array represents JSON items list
type Array struct {
	baseValue

	// Length is array length
	Length int
	// Items contains items list
	Items []Value
}

func newArray(pos Position, items ...Value) *Array {
	return &Array{
		baseValue: baseValue{pos},
		Length:    len(items),
		Items:     items,
	}
}

// Type implements jsonx.Value
func (_ Array) Type() Type {
	return TypeArray
}

// Interface implements json.Value
func (a Array) Interface() interface{} {
	out := make([]interface{}, 0, len(a.Items))
	for _, v := range a.Items {
		out = append(out, v.Interface())
	}
	return out
}

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

// Type implements jsonx.Value
func (_ Object) Type() Type {
	return TypeObject
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

// Null is JSON null value
type Null struct {
	baseValue
}

// Type implements jsonx.Value
func (_ Null) Type() Type {
	return TypeNull
}

func newNull(pos Position) Null {
	return Null{baseValue{pos}}
}

// Interface() implements json.Value
func (n Null) Interface() interface{} {
	return nil
}
