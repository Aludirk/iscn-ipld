package data

import (
	"fmt"
)

// ==================================================
// String
// ==================================================

// String is a data handler for the string
type String struct {
	*Base

	value  string
	filter *map[string]struct{}
}

var _ Data = (*String)(nil)

// NewString creates a string data handler
func NewString(key string, isRequired bool) *String {
	return &String{
		Base: NewBase(key, isRequired),
	}
}

// NewStringWithFilter creates a string data handler with filter
func NewStringWithFilter(key string, isRequired bool, filter []string) *String {
	filterPtr := &map[string]struct{}{}
	for _, value := range filter {
		(*filterPtr)[value] = struct{}{}
	}

	return &String{
		Base:   NewBase(key, isRequired),
		filter: filterPtr,
	}
}

// Prototype creates a prototype String
func (d *String) Prototype() Data {
	return &String{
		Base:   d.Base.Prototype(),
		filter: d.filter,
	}
}

// Get returns the string value
func (d *String) Get() string {
	return d.value
}

// Set the value of String
func (d *String) Set(obj interface{}) error {
	if value, ok := obj.(string); ok {
		if d.filter != nil {
			if _, ok := (*d.filter)[value]; !ok {
				return fmt.Errorf("String: %q is not a valid value", value)
			}
		}

		d.value = value
		d.Base.MarkDefined()
		return nil
	}

	return fmt.Errorf("String: 'string' is expected but '%T' is found", obj)
}

// Encode String
func (d *String) Encode() (interface{}, error) {
	return d.value, nil
}

// Decode String
func (d *String) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.value, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *String) ToJSON() (interface{}, error) {
	return d.value, nil
}

// Resolve resolves the value
func (d *String) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.value, nil, nil
}
