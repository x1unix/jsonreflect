package jsonx

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseNumber parses string into jsonx.Number
func ParseNumber(pos Position, str string, bitSize int) (*Number, error) {
	if str == "" || str == "0" {
		return &Number{baseValue: baseValue{pos}}, nil
	}

	// strconv.ParseFloat is not precise enough
	chunks := strings.SplitN(str, ".", 2)
	mantissaPart := chunks[0]
	isNegative := mantissaPart[0] == '-'
	mantissa, err := strconv.ParseInt(mantissaPart, 10, bitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mantissa part of number (%w)", err)
	}

	if len(chunks) < 2 {
		return &Number{
			baseValue: baseValue{pos},
			mantissa:  mantissa,
			IsSigned:  isNegative,
		}, nil
	}

	expoLen := len(chunks[1])
	exponent, err := strconv.ParseUint(chunks[1], 10, bitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exponent part of number (%w)", err)
	}

	return &Number{
		baseValue: baseValue{pos},
		IsFloat:   true,
		IsSigned:  isNegative,

		mantissa: mantissa,
		exponent: exponent,
		expoLen:  expoLen,
	}, nil
}
