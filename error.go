package jsonreflect

import (
	"errors"
	"fmt"
)

var (
	// ErrNotStringable means that value cannot be converted to string representation.
	ErrNotStringable = errors.New("value not stringable")
)

type ParseError struct {
	Position

	Message string
}

func NewParseError(pos Position, msg string, args ...interface{}) ParseError {
	err := ParseError{Position: pos}
	if len(args) > 0 {
		err.Message = fmt.Sprintf(msg, args...)
	} else {
		err.Message = msg
	}

	return err
}

func (p ParseError) Error() string {
	return fmt.Sprintf("%s (in range %d:%d)", p.Message, p.Start, p.End)
}

func NewUnexpectedCharacterError(start, end int, char byte) ParseError {
	return NewParseError(newPosition(start, end), "unexpected character %q", string(char))
}

func NewInvalidExprError(start, end int, val []byte) ParseError {
	return NewParseError(newPosition(start, end), "unexpected %q", string(val))
}
