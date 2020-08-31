package jsonx

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/x1unix/go-jsonx/internal/testutil"
)

func TestNewParser(t *testing.T) {
	input := []byte("foo")
	p := NewParser(input)
	require.Equal(t, len(input), p.end)
}

func TestParser_Parse(t *testing.T) {
	cases := map[string]struct {
		skip    bool
		src     FixtureProvider
		wantErr ExpectedError
		want    Value
	}{
		"empty document": {
			//skip: true
			src:  FixtureFromString("\t\r\n "),
			want: nil,
		},
		"single int number": {
			//skip: true,
			src: FixtureFromString("1024"),
			want: &Number{
				baseValue: newBaseValue(0, 3),
				mantissa:  1024,
			},
		},
		"single int number with padding": {
			//skip: true,
			src: FixtureFromString("\t\n1024\n"),
			want: &Number{
				baseValue: newBaseValue(2, 5),
				mantissa:  1024,
			},
		},
		"single float": {
			//skip: true,
			src: FixtureFromString("10.24"),
			want: &Number{
				baseValue: newBaseValue(0, 4),
				expoLen:   2,
				mantissa:  10,
				exponent:  24,
				IsFloat:   true,
			},
		},
		"single float with padding": {
			//skip: true,
			src: FixtureFromString("\t10.24 "),
			want: &Number{
				baseValue: newBaseValue(1, 5),
				expoLen:   2,
				mantissa:  10,
				exponent:  24,
				IsFloat:   true,
			},
		},
		"negative float": {
			//skip: true,
			src: FixtureFromString("-10.24"),
			want: &Number{
				baseValue: newBaseValue(0, 5),
				expoLen:   2,
				mantissa:  -10,
				exponent:  24,
				IsFloat:   true,
				IsSigned:  true,
			},
		},
		"invalid float with multiple dots": {
			src:     FixtureFromString(" 10.20.30 "),
			wantErr: `unexpected "10.20.30" (in range 1:9)`,
		},
		"invalid negative float with multiple negative chars": {
			src:     FixtureFromString(" ----10"),
			wantErr: `unexpected "----10" (in range 1:7)`,
		},
		"invalid number": {
			//skip:    true,
			src:     FixtureFromString("\t10fuu"),
			wantErr: `unexpected "10fuu"`,
		},
		"trash after number": {
			//skip:    true,
			src:     FixtureFromString("\t10 fuu"),
			wantErr: `unexpected "fuu"`,
		},
		"single true boolean": {
			//skip: true,
			src:  FixtureFromString(" true\t"),
			want: newBoolean(newPosition(1, 4), true),
		},
		"not bool but starts with bool expr": {
			//skip:    true,
			src:     FixtureFromString(" falsebuttrue\t"),
			wantErr: `unexpected "falsebuttrue"`,
		},
		"single false boolean": {
			//skip: true,
			src:  FixtureFromString("\n false\t "),
			want: newBoolean(newPosition(2, 6), false),
		},
		"single bad boolean": {
			//skip:    true,
			src:     FixtureFromString("\n boo\t "),
			wantErr: `unexpected character "b"`,
		},
		"same length and prefix as scalar but not scalar": {
			//skip:    true,
			src:     FixtureFromString("\ntruf\t "),
			wantErr: `unexpected "truf"`,
		},
		"single null": {
			//skip: true,
			src:  FixtureFromString("\n null\t "),
			want: newNull(newPosition(2, 5)),
		},
		"incomplete null": {
			//skip:    true,
			src:     FixtureFromString("nu"),
			wantErr: `unexpected "nu"`,
		},
		"invalid expr but contains null": {
			//skip:    true,
			src:     FixtureFromString("nullbutinvalid"),
			wantErr: `unexpected "nullbutinvalid"`,
		},
		"single string": {
			//skip: true,
			src:  FixtureFromString("\t\"foo\\nbar\\\\baz\"\n"),
			want: newString(newPosition(1, 15), []byte(`"foo\nbar\\baz"`)),
		},
		"unterminated single string": {
			//skip: true,
			src:     FixtureFromString("\t\"foo\\nbar"),
			wantErr: `unterminated string '"foo\nbar'`,
		},
		"trash after valid contents": {
			//skip: true,
			src:     FixtureFromString(`"foo",abcd`),
			wantErr: `unexpected ",abcd"`,
		},
		"empty array": {
			src:  FixtureFromString("[]"),
			want: newArray(newPosition(0, 1)),
		},
		"empty array with paddings": {
			src:  FixtureFromString("[\t\n ]"),
			want: newArray(newPosition(0, 4)),
		},
		"unterminated array": {
			src:     FixtureFromString("[\t\n true"),
			wantErr: `unterminated array statement (in range 0:8)`,
		},
		"array with trailing comma": {
			src:     FixtureFromString("[\t\n true ,]"),
			wantErr: `unexpected character "," (in range 9:10)`,
		},
		"simple array": {
			src: FixtureFromString(`[true]`),
			want: newArray(newPosition(0, 5),
				newBoolean(newPosition(1, 4), true)),
		},
		"array of scalar values": {
			//skip: true,
			src: TestdataFixture("arr_scalar.json"),
			want: newArray(newPosition(0, 32),
				newBoolean(newPosition(1, 4), true),
				newBoolean(newPosition(7, 11), false),
				newString(newPosition(14, 18), []byte(`"foo"`)),
				&Number{
					baseValue: newBaseValue(21, 24),
					mantissa:  32,
					exponent:  2,
					expoLen:   1,
					IsFloat:   true,
				},
				newNull(newPosition(27, 30))),
		},
		"empty object": {
			src:  FixtureFromString("{}"),
			want: newObject(0, 1, map[string]Value{}),
		},
		"object - unterminated prop name": {
			src:     FixtureFromString(`{"foo`),
			wantErr: ExpectedError(`unterminated string '"foo' (in range 1:5)`),
		},
		"object - trailing comma": {
			src:     FixtureFromString(`{"foo": 32,"bar":32,}`),
			wantErr: ExpectedError(`unexpected character "}" (in range 19:20)`),
		},
		"object - unterminated": {
			src:     FixtureFromString(`{"foo": 32,"bar":32`),
			wantErr: ExpectedError(`unterminated object (in range 0:19)`),
		},
		"object - invalid value separator": {
			src:     FixtureFromString(`{"foo":"bar",,`),
			wantErr: ExpectedError(`unexpected character "," (in range 0:13)`),
		},
		"object - invalid key-value separator": {
			src:     FixtureFromString(`{"foo"-32}`),
			wantErr: ExpectedError(`unexpected "-" (in range 0:6)`),
		},
		"object - non string literal key": {
			src:     FixtureFromString(`{10: 32}`),
			wantErr: ExpectedError(`unexpected character "1" (in range 0:1)`),
		},
		"object - invalid string literal key": {
			src:     FixtureFromString(`{"\c": 32}`),
			wantErr: ExpectedError(`jsonx.String: failed to unquote raw string value '"\c"': invalid syntax (in range 0:1)`),
		},
		"object - unterminated with padding": {
			src:     FixtureFromString("{\"foo\":\t\n"),
			wantErr: ExpectedError(`unterminated object (in range 0:7)`),
		},
		"invalid object value": {
			src:     FixtureFromString(`{"foo": fals}`),
			wantErr: ExpectedError(`unexpected "fals" (in range 8:12)`),
		},
		"object with one prop": {
			src: FixtureFromString(`{"foo": 10}`),
			want: newObject(0, 10, map[string]Value{
				"foo": &Number{
					baseValue: newBaseValue(8, 9),
					mantissa:  10,
				},
			}),
		},
		"object with two prop": {
			src: FixtureFromString(`{"foo": 10,"bar":true}`),
			want: newObject(0, 21, map[string]Value{
				"foo": &Number{
					baseValue: newBaseValue(8, 9),
					mantissa:  10,
				},
				"bar": newBoolean(newPosition(17, 20), true),
			}),
		},
		"nested object": {
			src: TestdataFixture("obj_nested.json"),
			want: newObject(0, 34, map[string]Value{
				"foo": newObject(11, 32, map[string]Value{
					"bar": newString(newPosition(24, 28), []byte(`"baz"`)),
				}),
			}),
		},
		"plain object with values": {
			src: TestdataFixture("obj_simple.json"),
			want: newObject(0, 288, map[string]Value{
				"id": &Number{
					baseValue: newBaseValue(10, 11),
					mantissa:  10,
				},
				"user": newString(newPosition(24, 30), []byte(`"admin"`)),
				"age": &Number{
					baseValue: newBaseValue(42, 43),
					mantissa:  32,
				},
				"created_at": newString(newPosition(62, 83), []byte(`"2009-11-10T23:00:00Z"`)),
				"roles": newArray(newPosition(97, 113),
					newString(newPosition(98, 103), []byte(`"root"`)),
					newString(newPosition(106, 112), []byte(`"owner"`))),
				"active": newBoolean(newPosition(128, 131), true),
				"rating": &Number{
					baseValue: newBaseValue(146, 152),
					mantissa:  -3,
					expoLen:   4,
					exponent:  1415,
					IsFloat:   true,
					IsSigned:  true,
				},
				"ref":         newNull(newPosition(164, 167)),
				"x-meta-salt": newString(newPosition(187, 220), []byte(`"d3b07384d113edec49eaa6238ad5ff00"`)),
				"meta": newObject(233, 286, map[string]Value{
					"first_name": newString(newPosition(253, 258), []byte(`"John"`)),
					"last_name":  newString(newPosition(278, 282), []byte(`"Doe"`)),
				}),
			}),
		},
	}

	tName, ok := IsOnlySubTest()
	if ok {
		_, ok := cases[tName]
		if !ok {
			t.Skipf("No such table test %q", tName)
		}
	}

	for n, c := range cases {
		if ok && tName != n {
			continue
		}
		if c.skip {
			continue
		}
		t.Run(n, func(t *testing.T) {
			src := c.src.ProvideFixture(t)
			got, err := NewParser(src).Parse()
			if !c.wantErr.AssertError(t, err) {
				require.Nil(t, got)
				return
			}

			require.Equal(t, c.want, got)
		})
	}
}
