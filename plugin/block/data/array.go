package data

import (
	"fmt"
	"reflect"
	"strconv"
)

// ==================================================
// Array
// ==================================================

// Array is an array of data handler
type Array struct {
	*Base

	array     []Data
	prototype Data
}

var _ Data = (*Array)(nil)

// NewDataArray creates an array of data handler
func NewDataArray(key string, isRequired bool, prototype Data) *Array {
	return &Array{
		Base:      NewBase(key, isRequired),
		array:     []Data{},
		prototype: prototype,
	}
}

// Prototype creates a prototype Array
func (d *Array) Prototype() Data {
	return &Array{
		Base:      d.Base.Prototype(),
		array:     []Data{},
		prototype: d.prototype.Prototype(),
	}
}

// Set the value of data handler array
func (d *Array) Set(obj interface{}) error {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(obj)
		for i := 0; i < s.Len(); i++ {
			elem := d.prototype.Prototype()
			if err := elem.Set(s.Index(i).Interface()); err != nil {
				return fmt.Errorf("(Index %d) %s", i, err.Error())
			}
			d.array = append(d.array, elem)
		}

		d.Base.MarkDefined()
		return nil
	}

	return fmt.Errorf("Array: an array is expected but '%T' is found", obj)
}

// Encode Array
func (d *Array) Encode() (interface{}, error) {
	res := []interface{}{}
	for i, obj := range d.array {
		enc, err := obj.Encode()
		if err != nil {
			return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
		}
		res = append(res, enc)
	}

	return res, nil
}

// Decode Array
func (d *Array) Decode(obj interface{}) (interface{}, error) {
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Slice:
		res := []interface{}{}
		s := reflect.ValueOf(obj)
		for i := 0; i < s.Len(); i++ {
			elem := d.prototype.Prototype()
			dec, err := elem.Decode(s.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
			}

			res = append(res, dec)
			d.array = append(d.array, elem)
		}

		d.Base.MarkDefined()
		return res, nil
	}

	return nil, fmt.Errorf("Array: an array is expected but '%T' is found", obj)
}

// ToJSON prepares the data for MarshalJSON
func (d *Array) ToJSON() (interface{}, error) {
	res := []interface{}{}
	for i, obj := range d.array {
		value, err := obj.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
		}
		res = append(res, value)
	}

	return res, nil
}

// Resolve resolves the value
func (d *Array) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		res := []interface{}{}
		for i, obj := range d.array {
			value, _, err := obj.Resolve(path)
			if err != nil {
				return nil, nil, fmt.Errorf("(Index %d) %s", i, err.Error())
			}
			res = append(res, value)
		}
		return res, nil, nil
	}

	first, rest := path[0], path[1:]
	index, err := strconv.ParseUint(first, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	if index >= uint64(len(d.array)) {
		return nil, nil, fmt.Errorf("index %d does not exist", index)
	}

	return d.array[index].Resolve(rest)
}
