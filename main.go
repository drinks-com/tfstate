package tfstate

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform/terraform"
)

// State is the interface for working with terraform state
type State interface {
	Read(workspace ...string) (*terraform.State, error)
	Write(ts *terraform.State) error
	Persist() error
}

// StructToMap converts a struct to a map using the struct's tags.
// It uses a 'map' tag on struct fields to decide which fields to add to the
// returned map.
func StructToMap(in interface{}) (map[string]interface{}, error) {
	tag := "map"
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("StructToMap only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			out[tagv] = v.Field(i).Interface()
		}
	}
	return out, nil
}
