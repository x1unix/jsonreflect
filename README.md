# jsonreflect

Package provides reflection features for JSON values.

## Goal

This package provides reflection features to JSON, which allows working with JSON structure
in `reflect` package manner.

Package might be useful in cases when unmarshal logic depends on source data structure like:

 * Object field type may vary (field can be array or object).
 * Object may contain fields stored in separate properties with same prefix, instead of array.
 * Object structure is polymorphic.

And for other cases of bad JSON structure design.

## Example

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
package myapi

import (
	"fmt"
	
	"github.com/x1unix/jsonreflect"
)

type Status struct {
	// status struct from json above
}

// checkStatus function checks if one of statuses contains error
func checkStatus(statuses ...Status) error

// checkResponseError checks if response has an error
func checkResponseError(resp []byte) error {
	// Check if response has error
    value, err := jsonreflect.NewParser(resp).Parse()
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
    switch jsonreflect.TypeOf(statusVal) {
    case jsonreflect.TypeArray:
        var statuses []Status
        if err = jsonreflect.UnmarshalValue(statusVal, &statuses); err != nil {
            return err
        }
        
        // perform check
        return checkStatus(statuses...)
    case jsonreflect.TypeObject:
    	status := &Status{}
    	if err = jsonreflect.UnmarshalValue(statusVal, &status); err != nil {
    		return err
        }
        
        // perform check
        return checkStatus(*status)
    default:
        // handle other cases
        return fmt.Errorf("unknown status type: %v", statusVal.Interface())
    }
}
```