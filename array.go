package jsonreflect

import "io"

// Array represents JSON items list
type Array struct {
	baseValue

	// Length is array length
	Length int
	// Items contains items list
	Items []Value
}

func newArray(pos Position, items ...Value) *Array {
	return &Array{
		baseValue: baseValue{pos},
		Length:    len(items),
		Items:     items,
	}
}

func (arr Array) marshal(w io.Writer, mf *marshalFormatter) error {
	if len(arr.Items) == 0 {
		return mf.write(w, []byte{tokenArrayStart, tokenArrayClose})
	}

	err := mf.writeOpenClause(w, tokenArrayStart)
	if err != nil {
		return err
	}

	childFmt := mf.childFormatter()
	lastIndex := len(arr.Items) - 1
	for i, v := range arr.Items {
		if err = childFmt.writePrefix(w); err != nil {
			return err
		}

		err = v.marshal(w, childFmt)
		if err != nil {
			return err
		}

		err = mf.writeElementDelimiter(w, i == lastIndex)
		if err != nil {
			return err
		}
	}
	return mf.write(w, []byte{tokenArrayClose})
}

// Type implements jsonreflect.Value
func (_ Array) Type() Type {
	return TypeArray
}

// Interface implements json.Value
func (arr Array) Interface() interface{} {
	out := make([]interface{}, 0, len(arr.Items))
	for _, v := range arr.Items {
		out = append(out, v.Interface())
	}
	return out
}
