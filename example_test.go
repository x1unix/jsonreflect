package jsonx

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

func ExampleParser_Parse() {
	// Plain JSON document parsing
	data, err := ioutil.ReadFile("testdata/obj_simple.json")
	must(err)

	// or NewParserFromReader() to read from io.Reader
	result, err := NewParser(data).Parse()
	must(err)

	// every json.Value has .Interface() method
	fmt.Println(result.Interface())

	// iterate over object
	obj := result.(*Object)
	for k, v := range obj.Items {
		fmt.Printf("- %s %T: %v\n", k, v, v.Interface())
	}

	// or simply return map
	fmt.Println(obj.ToMap())
}

func ExampleObject_GroupNumericKeys() {
	// Group similar keys by numeric suffix.
	// Suffix will be used as array index.
	//
	// see more examples in keys_test.go
	re := regexp.MustCompile(`^fan([\d]+)?$`)
	src := []byte(`{"fan3": 310, "fan1": 110, "fan2": 210, "foo": "bar"}`)
	doc, _ := NewParser(src).Parse()

	// Group keys by "fan[0-9]" pattern with regex and only 1 group match
	obj := doc.(*Object)
	grouped, _ := obj.GroupNumericKeys(re, 1)
	for _, item := range grouped {
		fmt.Println(item.Order, item.Key, obj.Items[item.Key].Interface())
	}
	// Output:
	// [1] fan1 110
	// [2] fan2 210
	// [3] fan3 310
}

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}
