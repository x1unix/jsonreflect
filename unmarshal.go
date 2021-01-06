package jsonreflect

import "fmt"

//import (
//	"encoding/json"
//	"fmt"
//	"reflect"
//)
//
//// Unmarshaler is the interface implemented by types that can unmarshal a JSON value description of themselves.
//type Unmarshaler interface {
//	UnmarshalJSONValue(v Value) error
//}
//
//func tryCallUnmarshaler(v Value, dst reflect.Value) (bool, error) {
//	if !dst.CanInterface() {
//		return false, nil
//	}
//
//	switch t := v.Interface().(type) {
//	case json.Unmarshaler:
//		return false, nil
//	}
//}
//
//func UnmarshalValue(v Value, dst interface{}) error {
//	dstVal := reflect.ValueOf(dst)
//	if dstVal.Kind() != reflect.Ptr {
//		return fmt.Errorf("passed value should be a pointer but got %s", dstVal.Type())
//	}
//
//}

func UnmarshalValue(v Value, dst interface{}) error {
	return fmt.Errorf("unimplemented!")
}
