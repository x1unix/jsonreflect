package jsonx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/x1unix/go-jsonx/internal/testutil"
)

func TestBaseValue_Ref(t *testing.T) {
	want := newPosition(1, 2)
	v := baseValue{Position: want}
	require.Equal(t, want, v.Ref())
}

func TestString_RawString(t *testing.T) {
	want := `"foo"`
	str := String{rawValue: []byte(want)}
	require.Equal(t, str.RawString(), want)
}

func TestString_String(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
		err  ExpectedError
	}{
		"valid quoted string": {
			in:   `"foo\nbar"`,
			want: "foo\nbar",
		},
		"invalid string": {
			in:  "foo",
			err: "failed to unquote raw string value 'foo': invalid syntax",
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			str := String{rawValue: []byte(c.in)}
			got, err := str.String()
			if !c.err.AssertError(t, err) {
				return
			}
			require.Equal(t, c.want, got)
		})
	}
}

func TestString_Number(t *testing.T) {
	cases := map[string]struct {
		in   string
		want *Number
		err  ExpectedError
	}{
		"valid quoted number": {
			in: `"3.14"`,
			want: &Number{
				baseValue: newBaseValue(0, 5),
				mantissa:  3,
				exponent:  14,
				expoLen:   2,
				IsFloat:   true,
			},
		},
		"quoted NaN": {
			in:  `"nan"`,
			err: "failed to parse mantissa part of number",
		},
		"unquoted": {
			in:  `1010`,
			err: "failed to unquote raw string value '1010'",
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			str := String{rawValue: []byte(c.in)}
			if c.want != nil {
				str.baseValue = c.want.baseValue
			}

			got, err := str.Number()
			if !c.err.AssertError(t, err) {
				return
			}
			require.Equal(t, c.want, got)
		})
	}
}

func TestString_Interface(t *testing.T) {
	cases := map[string]struct {
		in   string
		want interface{}
		err  ExpectedError
	}{
		"valid quoted string": {
			in:   `"foo\nbar"`,
			want: "foo\nbar",
		},
		"invalid string": {
			in:   "foo",
			want: "foo",
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			str := String{rawValue: []byte(c.in)}
			got := str.Interface()
			require.NotNil(t, got)
			require.Equal(t, c.want, got)
		})
	}
}

func TestBoolean_Interface(t *testing.T) {
	b := Boolean{Value: true}
	require.Equal(t, true, b.Interface())
}

func TestNumber_Float32_and_64(t *testing.T) {
	cases := map[float64]*Number{
		32: {
			mantissa: 32,
		},
		3.14: {
			mantissa: 3,
			exponent: 14,
			expoLen:  2,
			IsFloat:  true,
		},
		-3.14: {
			mantissa: -3,
			exponent: 14,
			expoLen:  2,
			IsFloat:  true,
		},
	}

	for want, num := range cases {
		t.Run(fmt.Sprintf("%v", want), func(t *testing.T) {
			got := num.Float64()
			require.Equal(t, want, got)
			require.Equal(t, float32(want), num.Float32())
		})
	}
}

func TestNumber_Int(t *testing.T) {
	in := 32
	n := Number{mantissa: 32}
	require.Equal(t, in, n.Int())
	require.Equal(t, int32(in), n.Int32())
	require.Equal(t, int64(in), n.Int64())
}

func TestNumber_Uint(t *testing.T) {
	in := uint(32)
	n := Number{mantissa: 32}
	require.Equal(t, in, n.Uint())
	require.Equal(t, uint32(in), n.Uint32())
	require.Equal(t, uint64(in), n.Uint64())
}

func TestNumber_Interface(t *testing.T) {
	n1 := Number{IsFloat: true, mantissa: 3, expoLen: 2, exponent: 14}
	require.Equal(t, 3.14, n1.Interface())
	n2 := Number{mantissa: 2}
	require.Equal(t, 2, n2.Interface())
}
