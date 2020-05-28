package data

import (
	"fmt"
)

// ==================================================
// FilterString
// ==================================================

// FilterString is a data handler for a string with filter
type FilterString struct {
	*Base

	value  *String
	filter map[string]struct{}
}

var _Data = (*FilterString)(nil)

// NewFilterString creates a string data handler with filter
func NewFilterString(key string, isRequired bool, filterList []string) *FilterString {
	filter := map[string]struct{}{}
	for _, value := range filterList {
		filter[value] = struct{}{}
	}

	return &FilterString{
		Base:   NewBase(key, isRequired),
		value:  NewString("", false),
		filter: filter,
	}
}

// Prototype creates a prototype FilterString
func (d *FilterString) Prototype() Data {
	return &FilterString{
		Base:   d.Base.Prototype(),
		value:  NewString(d.value.GetKey(), d.value.IsRequired()),
		filter: d.filter,
	}
}

// Get returns the string value
func (d *FilterString) Get() string {
	return d.value.Get()
}

// Set the value of FilterString string
func (d *FilterString) Set(obj interface{}) error {
	if err := d.value.Set(obj); err != nil {
		return err
	}

	if _, ok := d.filter[d.value.Get()]; !ok {
		return fmt.Errorf("FilterString: %q is filtered out", d.value.Get())
	}

	d.Base.MarkDefined()
	return nil
}

// Encode FilterString
func (d *FilterString) Encode() (interface{}, error) {
	return d.value.Encode()
}

// Decode FilterString
func (d *FilterString) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.value.Get(), nil
}

// ToJSON prepares the data for MarshalJSON
func (d *FilterString) ToJSON() (interface{}, error) {
	return d.value.ToJSON()
}

// Resolve resolves the value
func (d *FilterString) Resolve(path []string) (interface{}, []string, error) {
	return d.value.Resolve(path)
}
