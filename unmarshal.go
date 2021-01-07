package jsonreflect

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

const (
	tagNameJSON = "json"

	tagOptionSkip          = "-"
	tagOptionCollectOrphan = "*"
)

var (
	typeJsonRawMessage = reflect.TypeOf((*json.RawMessage)(nil)).Elem
)

// Unmarshaler is the interface implemented by types that can unmarshal a JSON value description of themselves.
type Unmarshaler interface {
	UnmarshalJSONValue(v Value) error
}

type unmarshalParams struct {
	strict                      bool
	dangerouslySetPrivateFields bool
}

func newUnmarshalParams(opts []UnmarshalOption) unmarshalParams {
	p := unmarshalParams{strict: true}
	if len(opts) == 0 {
		return p
	}

	for _, opt := range opts {
		opt(&p)
	}
	return p
}

// UnmarshalOption is unmarshal option
type UnmarshalOption func(fn *unmarshalParams)

var (
	// NoStrict disables unmarshal strict mode.
	//
	// When strict mode is disabled, unmarshaler will try to cast JSON value
	// to destination value.
	//
	// Supported possible casts (without strict mode):
	//	// If destination type is numeric, unmarshaler will try
	//	// to parse source value as number.
	//	string -> any numeric value
	//
	//	// Any numeric value can be casted to string
	//  any numeric value -> string
	//
	//	// Any valid boolean can be casted to string (and vice versa)
	//	boolean <-> string
	//
	NoStrict UnmarshalOption = func(fn *unmarshalParams) {
		fn.strict = false
	}

	// DangerouslySetPrivateFields allows unmarshaler to modify private fields
	// which have `json` tag.
	//
	// Use it if you really know what to do, you have been warned.
	//
	// We are not responsible for corrupted memory, dead hard drives, thermonuclear war,
	// or you getting fired because the production database went down.
	//
	// Please do some research if you have any concerns about this option.
	DangerouslySetPrivateFields UnmarshalOption = func(fn *unmarshalParams) {
		fn.dangerouslySetPrivateFields = true
	}
)

func tryCallUnmarshaler(v Value, dst reflect.Value) (bool, error) {
	if !dst.CanInterface() {
		return false, nil
	}

	switch t := v.Interface().(type) {
	case json.Unmarshaler:
		str, err := MarshalValue(v, nil)
		if err != nil {
			return false, err
		}

		return true, t.UnmarshalJSON(str)
	case Unmarshaler:
		return true, t.UnmarshalJSONValue(v)
	case json.RawMessage:
		serialized, err := MarshalValue(v, nil)
		if err != nil {
			return false, err
		}

		dst.Set(reflect.ValueOf(serialized))
		return true, nil
	default:
		return false, nil
	}
}

// UnmarshalValue maps JSON value to passed value.
// Accepts additional options to customise unmarshal process.
//
// Method supports the same tag and behavior as standard json.Unmarshal method.
//
// Supported additional tags:
//
// - `json:"*"` tag used to collect all orphan values in JSON object to specified field.
//
// Supported special unmarshal types:
//
// - If destination value is jsonreflect.Value, unmarshaler will map original value.
//
// - If destination value is jsonreflect.Unmarshaler, unmarshaler will call Unmarshaler.UnmarshalJSONValue.
//
func UnmarshalValue(v Value, dst interface{}, opts ...UnmarshalOption) error {
	params := newUnmarshalParams(opts)
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("passed value should be a pointer but got %s", dstVal.Type())
	}

	if dstVal.IsNil() {
		return errors.New("nil pointer passed")
	}

	dstElem := dstVal.Elem()
	return unmarshalValue(v, dstElem, params)
}

func unmarshalValue(src Value, dst reflect.Value, p unmarshalParams) error {
	if !dst.CanSet() {
		return errors.New("destination value must be exported")
	}

	isUnmarshed, err := tryCallUnmarshaler(src, dst)
	if err != nil {
		return err
	}

	if isUnmarshed {
		return nil
	}

	dstType := dst.Type()
	if dst.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
		if dst.IsNil() {
			// initialize pointer value
			dst.Set(reflect.New(dstType))
		}

		// unref pointer
		dst = dst.Elem()
	}

	switch k := dstType.Kind(); k {
	case reflect.String:
		return unmarshalString(src, dst, p.strict)
	case reflect.Bool:
		return unmarshalBool(src, dst, p.strict)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return unmarshalUint(src, dst, p.strict)
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return unmarshalInt(src, dst, p.strict)
	case reflect.Float32, reflect.Float64:
		return unmarshalFloat(src, dst, p.strict)
	case reflect.Slice:
		return unmarshalSlice(src, dst, p)
	case reflect.Array:
		return unmarshalArray(src, dst, p)
	case reflect.Map:
		return unmarshalMap(src, dst, p)
	case reflect.Struct:
		return unmarshalObject(src, dst, p)
	case reflect.Interface:
		return unmarshalInterface(src, dst)
	}

	return nil
}

type tagData struct {
	skipValue      bool
	collectOrphans bool
	srcKey         string
}

func parseTagData(f reflect.StructField) *tagData {
	val, ok := f.Tag.Lookup(tagNameJSON)
	if !ok {
		return nil
	}

	if val == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	if parts[0] == tagOptionSkip {
		return &tagData{skipValue: true}
	}

	srcKey := strings.TrimSpace(parts[0])
	switch srcKey {
	case "":
		return nil
	case tagOptionCollectOrphan:
		return &tagData{collectOrphans: true}
	default:
		return &tagData{srcKey: srcKey}
	}
}

// findSourceKey attempts to find source object key to unmarshal.
//
// First it tries to find `json` tag declaration.
// If no tag available, method tries to find source key using property name with different cases.
func findSourceKey(td *tagData, srcObj *Object, fType reflect.StructField) (string, bool) {
	if td != nil && td.srcKey != "" {
		if srcObj.HasKey(td.srcKey) {
			return td.srcKey, true
		}

		return "", false
	}

	if srcObj.HasKey(fType.Name) {
		return fType.Name, true
	}

	// try to cast to camel case and lookup
	ccName := strcase.ToLowerCamel(fType.Name)
	if srcObj.HasKey(ccName) {
		return ccName, true
	}

	return "", false
}

func unmarshalObject(src Value, dst reflect.Value, p unmarshalParams) error {
	srcObj, ok := src.(*Object)
	if !ok {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	// orphan keys registry
	touchedKeys := make(map[string]struct{})
	var orphanDest *reflect.Value

	for i := 0; i < dst.NumField(); i++ {
		fType := dst.Type().Field(i)
		fVal := dst.Field(i)

		tagData := parseTagData(fType)
		if tagData != nil && tagData.skipValue {
			continue
		}

		if !fVal.CanSet() {
			// DangerouslySetPrivateFields() option captures private fields with valid `json` tag.
			if !(p.dangerouslySetPrivateFields && tagData != nil) {
				continue
			}

			// Here be dragons 🔥
			fVal = reflect.NewAt(fVal.Type(), unsafe.Pointer(fVal.UnsafeAddr())).Elem()
		}

		if !fType.Anonymous {
			// mark value as target for all orphan values
			// if it has `json:"*"` tag.
			if tagData != nil && tagData.collectOrphans {
				orphanDest = &fVal
				continue
			}

			srcKey, ok := findSourceKey(tagData, srcObj, fType)
			if !ok {
				continue
			}

			touchedKeys[srcKey] = struct{}{}
			srcVal := srcObj.Items[srcKey]
			if err := unmarshalValue(srcVal, fVal, p); err != nil {
				return fmt.Errorf("can't unmarshal field %q to %s.%s: %w", srcKey, dst.Type(), fType.Type, err)
			}
			continue
		}

		if err := unmarshalValue(src, fVal, p); err != nil {
			return fmt.Errorf("can't unmarshal to %s.%s: %w", dst.Type(), fType.Type, err)
		}
		continue
	}

	if orphanDest == nil {
		return nil
	}

	// unmarshal orphan values (if requested)
	if err := unmarshalOrphanKeys(srcObj, touchedKeys, *orphanDest, p); err != nil {
		return fmt.Errorf("failed to unmarshal orphan keys to %s: %w", orphanDest.Type(), err)
	}
	return nil
}

func unmarshalOrphanKeys(srcObj *Object, touchedKeys map[string]struct{}, dst reflect.Value, p unmarshalParams) error {
	orphans := make(map[string]Value)
	for k, v := range srcObj.Items {
		if _, ok := touchedKeys[k]; ok {
			continue
		}

		orphans[k] = v
	}

	orphansContainer := &Object{
		baseValue: srcObj.baseValue,
		Items:     orphans,
	}
	return unmarshalValue(orphansContainer, dst, p)
}

func unmarshalInterface(src Value, dst reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot assign %s to %s: %v", src.Type(), dst.Type(), r)
		}
	}()

	iface := reflect.ValueOf(src.Interface())
	dst.Set(iface)
	return nil
}

func unmarshalMap(src Value, dst reflect.Value, p unmarshalParams) error {
	srcObj, ok := src.(*Object)
	if !ok {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	if k := dst.Type().Key().Kind(); k != reflect.String {
		return fmt.Errorf("destination map key type should be string (got %s)", k)
	}

	elemType := dst.Type().Elem()
	m := reflect.MakeMap(dst.Type())
	for key, value := range srcObj.Items {
		newVal := reflect.New(elemType)
		if err := unmarshalValue(value, newVal, p); err != nil {
			return fmt.Errorf("%q: cannot set %s to map value: %w", key, src.Type(), err)
		}

		m.SetMapIndex(reflect.ValueOf(key), newVal.Elem())
	}

	dst.Set(m)
	return nil
}

func unmarshalArray(src Value, dst reflect.Value, p unmarshalParams) error {
	srcArr, ok := src.(*Array)
	if !ok {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	items := srcArr.Items
	maxLen := dst.Type().Len()
	if len(items) > maxLen {
		if p.strict {
			return fmt.Errorf("value overflows destination array (%d > %d)", len(items), maxLen)
		}

		// just trim array in unsafe mode
		items = items[:maxLen]
	}

	for i, val := range srcArr.Items {
		if err := unmarshalValue(val, dst.Index(i), p); err != nil {
			return fmt.Errorf("can't assign %s to index #%d: %w", val.Type(), i, err)
		}
	}

	return nil
}

func unmarshalSlice(src Value, dst reflect.Value, p unmarshalParams) error {
	srcArr, ok := src.(*Array)
	if !ok {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	arrLen := len(srcArr.Items)
	slice := reflect.MakeSlice(dst.Type(), arrLen, arrLen)
	for i, val := range srcArr.Items {
		if err := unmarshalValue(val, slice.Index(i), p); err != nil {
			return fmt.Errorf("can't set %s to index #%d: %w", val.Type(), i, err)
		}
	}

	dst.Set(slice)
	return nil
}

func unmarshalFloat(src Value, dst reflect.Value, strict bool) error {
	bitness := 64
	if dst.Kind() == reflect.Float32 {
		bitness = 32
	}

	if strict && src.Type() != TypeNumber {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	numval, err := ToNumber(src, bitness)
	if err != nil {
		return err
	}

	dst.SetFloat(numval.Float64())
	return nil
}

func unmarshalInt(src Value, dst reflect.Value, strict bool) error {
	if strict && src.Type() != TypeNumber {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	numval, err := ToNumber(src, 64)
	if err != nil {
		return err
	}

	dst.SetInt(numval.Int64())
	return nil
}

func unmarshalUint(src Value, dst reflect.Value, strict bool) error {
	if strict && src.Type() != TypeNumber {
		return newUnmarshalTypeErr(src.Type(), dst.Type())
	}

	numval, err := ToNumber(src, 64)
	if err != nil {
		return err
	}

	if numval.IsSigned {
		return fmt.Errorf("assignment of signed value %v to unsigned type %s", numval.Interface(), dst.Type())
	}

	dst.SetUint(numval.Uint64())
	return nil
}

func unmarshalBool(src Value, dst reflect.Value, strict bool) error {
	switch t := TypeOf(src); t {
	case TypeBoolean:
		dst.SetBool(src.(*Boolean).Value)
		return nil
	case TypeString:
		if strict {
			return newUnmarshalTypeErr(t, dst.Type())
		}

		strval, err := src.String()
		if err != nil {
			return err
		}

		boolval, err := strconv.ParseBool(strval)
		if err != nil {
			return newUnmarshalCastErr(t, dst.Type(), err)
		}

		dst.SetBool(boolval)
		return nil
	default:
		return newUnmarshalTypeErr(t, dst.Type())
	}
}

func unmarshalString(src Value, dst reflect.Value, strict bool) error {
	if t := TypeOf(src); strict && t != TypeString {
		return newUnmarshalTypeErr(t, dst.Type())
	}

	strval, err := src.String()
	if err != nil {
		return newUnmarshalCastErr(src.Type(), dst.Type(), err)
	}

	dst.SetString(strval)
	return nil
}

func newUnmarshalTypeErr(srcType Type, dstType reflect.Type) error {
	return fmt.Errorf("cannot unmarshal %s value to %s", srcType, dstType)
}

func newUnmarshalCastErr(srcType Type, dstType reflect.Type, err error) error {
	return fmt.Errorf("cannot convert %s value to destination value %s: %w", srcType, dstType, err)
}
