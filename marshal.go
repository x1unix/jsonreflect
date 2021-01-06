package jsonx

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

const (
	charLineBreak = '\n'
	charSpace     = ' '
)

type marshalFormatter struct {
	isRoot bool
	indent []byte
	level  int
}

func (mf *marshalFormatter) writePrefix(w io.Writer) error {
	if mf.noIndent() {
		return nil
	}

	_, err := w.Write(bytes.Repeat(mf.indent, mf.level))
	return err
}

func (mf *marshalFormatter) writeString(w io.Writer, str string) error {
	return mf.write(w, []byte(str))
}

func (mf *marshalFormatter) noIndent() bool {
	return mf == nil || len(mf.indent) == 0
}

func (mf *marshalFormatter) writePropertyName(w io.Writer, name string) error {
	quotedName := []byte(strconv.Quote(name))
	if mf.noIndent() {
		_, err := w.Write(append(quotedName, tokenKeyDelimiter))
		return err
	}

	return mf.write(w, append(quotedName, tokenKeyDelimiter, charSpace))
}

func (mf *marshalFormatter) writeOpenClause(w io.Writer, chr byte) (err error) {
	if mf.noIndent() {
		_, err = w.Write([]byte{chr})
		return err
	}

	_, err = w.Write([]byte{chr, charLineBreak})
	return err
}

func (mf *marshalFormatter) writeElementDelimiter(w io.Writer, isLast bool) error {
	if mf.noIndent() {
		if isLast {
			return nil
		}
		_, err := w.Write([]byte{tokenDelimiter})
		return err
	}

	if isLast {
		_, err := w.Write([]byte{charLineBreak})
		return err
	}

	_, err := w.Write([]byte{tokenDelimiter, charLineBreak})
	return err
}

func (mf *marshalFormatter) write(w io.Writer, data []byte) error {
	if mf == nil {
		_, err := w.Write(data)
		return err
	}

	err := mf.writePrefix(w)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func (mf *marshalFormatter) childFormatter() *marshalFormatter {
	if mf == nil {
		return nil
	}
	return &marshalFormatter{isRoot: false, indent: mf.indent, level: mf.level + 1}
}

// MarshalOptions contains additional marshal options
type MarshalOptions struct {
	// Indent is indentation to apply for output
	Indent string
}

func (opts *MarshalOptions) formatter() *marshalFormatter {
	if opts == nil {
		return nil
	}

	return &marshalFormatter{
		isRoot: true,
		indent: []byte(opts.Indent),
	}
}

// MarshalValue returns the JSON encoding of passed jsonx.Value
//
// Accepts optional argument which allows to specify indent.
func MarshalValue(v Value, opts *MarshalOptions) ([]byte, error) {
	buff := &bytes.Buffer{}
	if err := v.marshal(buff, opts.formatter()); err != nil {
		return nil, fmt.Errorf("failed to marshal JSON %s: %w", v.Type(), err)
	}
	return buff.Bytes(), nil
}
