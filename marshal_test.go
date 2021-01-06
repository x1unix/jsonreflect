package jsonreflect

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalValue(t *testing.T) {
	cases := map[string]struct {
		srcFile string
		equalTo []byte
		opts    *MarshalOptions
	}{
		"foo": {
			srcFile: "test_marshal_value.json",
			opts:    &MarshalOptions{Indent: "  "},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			f, err := ioutil.ReadFile(filepath.Join("testdata", c.srcFile))
			require.NoError(t, err)

			val, err := NewParser(f).Parse()
			require.NoError(t, err, "json parse failed")

			got, err := MarshalValue(val, c.opts)
			require.NoError(t, err, "json marshal failed")

			if len(c.equalTo) == 0 {
				c.equalTo = f
			}

			require.Equal(t, string(c.equalTo), string(got))
		})
	}
}
