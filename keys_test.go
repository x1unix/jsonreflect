package jsonreflect

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/x1unix/jsonreflect/internal/testutil"
)

func TestObject_GroupNumericKeys(t *testing.T) {
	cases := map[string]struct {
		re        *regexp.Regexp
		src       FixtureProvider
		groupSize int
		want      GroupedNumbericKeys
		err       ExpectedError
	}{
		"one level group": {
			groupSize: 1,
			re:        regexp.MustCompile(`^fan([\d]+)?$`),
			src:       FixtureFromString(`{"fan3": 30, "fan1": 10, "fan2": 20, "foo": "bar"}`),
			want: GroupedNumbericKeys{
				{
					Order: []int{1},
					Key:   "fan1",
				},
				{
					Order: []int{2},
					Key:   "fan2",
				},
				{
					Order: []int{3},
					Key:   "fan3",
				},
			},
		},
		"multi-level group": {
			groupSize: 2,
			re:        regexp.MustCompile(`^temp([\d]+)[_]?([\d]+)?$`),
			src:       TestdataFixture("obj_key_numgroup.json"),
			want: GroupedNumbericKeys{
				{
					Order: []int{1, 0},
					Key:   "temp1",
				},
				{
					Order: []int{2, 0},
					Key:   "temp2",
				},
				{
					Order: []int{3, 0},
					Key:   "temp3",
				},
				{
					Order: []int{3, 1},
					Key:   "temp3_1",
				},
				{
					Order: []int{3, 2},
					Key:   "temp3_2",
				},
			},
		},
		"handle non-numeric keys": {
			groupSize: 1,
			re:        regexp.MustCompile(`^fan([A-Za-z0-9]+)?$`),
			src:       FixtureFromString(`{"fan3": 30, "fanA": 10}`),
			err:       `segment "A" of key "fanA" is not a number`,
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			td := c.src.ProvideFixture(t)
			v, err := NewParser(td).Parse()
			require.NoError(t, err)
			obj, ok := v.(*Object)
			require.True(t, ok)
			got, err := obj.GroupNumericKeys(c.re, c.groupSize)
			if !c.err.AssertError(t, err) {
				return
			}

			require.Equal(t, c.want, got)
		})
	}
}
