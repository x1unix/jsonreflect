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
		"single int number": {
			skip: true,
			src:  FixtureFromString("1024"),
			want: &Number{
				baseValue: newBaseValue(0, 3),
				mantissa:  1024,
			},
		},
		"single int number with padding": {
			skip: true,
			src:  FixtureFromString("\t\n1024\n"),
			want: &Number{
				baseValue: newBaseValue(2, 5),
				mantissa:  1024,
			},
		},
		"single float": {
			skip: true,
			src:  FixtureFromString("10.24"),
			want: &Number{
				baseValue: newBaseValue(0, 4),
				expoLen:   2,
				mantissa:  10,
				exponent:  24,
				IsFloat:   true,
			},
		},
		"single float with padding": {
			skip: true,
			src:  FixtureFromString("\t10.24 "),
			want: &Number{
				baseValue: newBaseValue(1, 5),
				expoLen:   2,
				mantissa:  10,
				exponent:  24,
				IsFloat:   true,
			},
		},
		"invalid number": {
			skip:    true,
			src:     FixtureFromString("\t10fuu"),
			wantErr: `unexpected character "f"`,
		},
		"single true boolean": {
			skip: true,
			src:  FixtureFromString(" true\t"),
			want: newBoolean(newPosition(1, 4), true),
		},
		"not bool but starts with bool expr": {
			// Fails now, needs refactor!!
			skip:    true,
			src:     FixtureFromString(" falsebuttrue\t"),
			wantErr: `unexpected character "f"`,
		},
		"single false boolean": {
			skip: true,
			src:  FixtureFromString("\n false\t "),
			want: newBoolean(newPosition(2, 6), false),
		},
		"single bad boolean": {
			skip:    true,
			src:     FixtureFromString("\n boo\t "),
			wantErr: `unexpected character "b"`,
		},
		"single null": {
			skip: true,
			src:  FixtureFromString("\n null\t "),
			want: newNull(newPosition(2, 5)),
		},
		"incomplete null": {
			skip:    true,
			src:     FixtureFromString("nu"),
			wantErr: `unexpected character "n"`,
		},
	}

	for n, c := range cases {
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

			require.NotNil(t, got)
			require.Equal(t, c.want, got)
		})
	}
}
