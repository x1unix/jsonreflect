package jsonreflect

import (
	"io"
	"math"
	"strconv"
	"strings"
)

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

// Type implements jsonreflect.Value
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

func (n Number) asString() string {
	if !n.IsFloat {
		return strconv.Itoa(n.Int())
	}
	sb := strings.Builder{}
	sb.WriteString(strconv.Itoa(n.Int()))
	sb.WriteRune('.')
	sb.WriteString(strconv.FormatUint(n.exponent, 10))
	return sb.String()
}

// String implements jsonreflect.Value
func (n Number) String() (string, error) {
	return n.asString(), nil
}

func (n Number) marshal(w io.Writer, _ *marshalFormatter) error {
	_, err := w.Write([]byte(n.asString()))
	return err
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
