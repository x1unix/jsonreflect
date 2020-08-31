// package testutil is internal test helpers package
package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// IsOnlySubTest returns name of single subtest that should be run
func IsOnlySubTest() (string, bool) {
	return os.LookupEnv("ONLY_SUBTEST")
}

// ExpectedError is string which should be in expected error
type ExpectedError string

// Empty returns if expected string is empty
func (x ExpectedError) Empty() bool {
	return x == ""
}

// AssertError checks if error contains expected string.
//
// Returns true if no errors expected or got and test can be continued.
//
// Example:
//	if !wantErr.AssertError(t, err) {
//		// Some error in test expected, so just exit
//		return
//	}
//	assert.Equal(t, want, got)
//
func (x ExpectedError) AssertError(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		if !x.Empty() {
			t.Fatalf("expected error which contains %q but got no errors", x)
		}
		return true
	}

	if x.Empty() {
		t.Fatalf("unexpected error: '%[1]s'\n\n\tRaw: %#[1]v", err)
	}

	if !strings.Contains(err.Error(), string(x)) {
		t.Fatalf("error '%s' should contain '%s'", err.Error(), x)
	}

	return false
}

// FixtureFromString indicates that source is plain string
type FixtureFromString string

// ProvideFixture implements FixtureProvider
func (j FixtureFromString) ProvideFixture(t *testing.T) []byte {
	return []byte(j)
}

// TestdataFixture indicates that source should be taken from file in testdata dir
type TestdataFixture string

// ProvideFixture implements FixtureProvider
func (j TestdataFixture) ProvideFixture(t *testing.T) []byte {
	f, err := os.Open(filepath.Join("testdata", string(j)))
	require.NoError(t, err, "ProvideFixture: cannot open fixture file")
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	require.NoError(t, err, "ProvideFixture: cannot read fixture file")
	return data
}

// FixtureProvider is test fixture provider
type FixtureProvider interface {
	ProvideFixture(t *testing.T) []byte
}
