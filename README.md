# jsonreflect

[![GoDoc](https://godoc.org/github.com/x1unix/jsonreflect?status.svg)](https://pkg.go.dev/github.com/x1unix/jsonreflect)

Package provides reflection features for JSON values.

## Goal

This package provides reflection features to JSON, which allows working with JSON structure
in `reflect` package manner.

Package might be useful in cases when unmarshal logic depends on source data structure like:

 * Object field type may vary (field can be array or object).
 * Object may contain fields stored in separate properties with same prefix, instead of array.
 * Object structure is polymorphic.

And for other cases of bad JSON structure design.

## Examples

### Processing unknown values

For instance, there is an object which only have a small subset of
known fields and there is a need to separate known values from orphans for future processing.

Example below shows how *jsonreflect* allows collecting all unknown fields on value unmarshal.

```go
package main

import (
    "fmt"
	
    "github.com/x1unix/jsonreflect"
)

type GenericResponse struct {
    // Known fields
    Status int      `json:"status"`
    Payload []byte  `json:"payload"`
    
    // Container for unknown fields.
    // Also, *json.Value can be used to get values as JSON object.
    Orphan map[string]interface{}  `json:"..."`
}

func unmarshalGenericResponse(data []byte) error {
    rsp := new(GenericResponse)
    if err := jsonreflect.Unmarshal(data, rsp); err != nil {
        return err
    }
    
    // Process orphan fields
    fmt.Println(rsp.Orphan)
    return nil
}

```

### Corrupted object

For instance, we have an JSON response from some service with specified structure,
but sometimes service returns response with different structure when internal error occurs.

**Normal response**
```json
{
  "STATUS": [
    {
      "STATUS": "S",
      "When": 1609265946,
      "Code": 9,
      "Msg": "3 ASC(s)"
    }
  ],
  "DATA": ["some actual data..."]
}
```

**Abnornal response:**

```json
{
  "STATUS": "E",
  "When": 1609267826,
  "Code": 14,
  "Msg": "invalid cmd",
  "Description": "cgminer v1.3"
}
```

In normal cases, `STATUS` is an array but sometimes it might be a regular object.
Let's mitigate this issue.

**Example**

```go
package main

import (
	"fmt"
	
	"github.com/x1unix/jsonreflect"
)

type Status struct {
	// status struct from json above
}

// checkStatus function checks if one of statuses contains error
func checkStatus(statuses []Status) error

// checkResponseError checks if response has an error
func checkResponseError(resp []byte) error {
    // Check if response has error
    value, err := jsonreflect.ValueOf(resp)
    if err != nil {
        // Invalid json
        return err
    }
    
    // cast response to object
    obj, err := jsonreflect.ToObject(value)
    if err != nil {
    	// handle invalid response
    	return fmt.Errorf("unexpected response: %v", value.Interface())
    }
    
    statusVal := obj.Items["STATUS"]
    
    // wrap status with array
    if jsonreflect.TypeOf(statusVal) != jsonreflect.TypeArray {
    	statusVal = jsonreflect.NewArray(statusVal)
    }
    
    // unmarshal value to struct and do our stuff
    var statuses []Status
    if err = jsonreflect.UnmarshalValue(statusVal, &statuses); err != nil {
    	return err
    }
    
    return checkStatus(statuses)
}
```
