package data

import (
	"fmt"

	"gitlab.com/c0b/go-ordered-json"
)

// ==================================================
// Object
// ==================================================

// Codec is a codec of an ISCN object
type Codec interface {
	MarkNested()

	GetData() (*ordered.OrderedMap, error)
	SetData(map[string]interface{}) error

	Encode() (map[string]interface{}, error)
	Decode(map[string]interface{}) error

	Resolve(path []string) (interface{}, []string, error)
}

// ObjectPrototypeFunc returns a factory function to create ISCN object prototype
type ObjectPrototypeFunc func() Codec

// Object is a data handler of a nested ISCN object
type Object struct {
	*Base

	prototypeFunc ObjectPrototypeFunc
	object        Codec
}

var _ Data = (*Object)(nil)

// NewObject creates a nested ISCN object data handler
func NewObject(key string, isRequired bool, prototypeFunc ObjectPrototypeFunc) *Object {
	object := prototypeFunc()
	object.MarkNested()

	return &Object{
		Base:          NewBase(key, isRequired),
		prototypeFunc: prototypeFunc,
		object:        object,
	}
}

// Prototype creates a prototype Object
func (d *Object) Prototype() Data {
	object := d.prototypeFunc()
	object.MarkNested()

	return &Object{
		Base:          d.Base.Prototype(),
		prototypeFunc: d.prototypeFunc,
		object:        object,
	}
}

// Set the value of Object
func (d *Object) Set(obj interface{}) error {
	if value, ok := obj.(map[string]interface{}); ok {
		if err := d.object.SetData(value); err != nil {
			return err
		}

		d.Base.MarkDefined()
		return nil
	}

	return fmt.Errorf("Object: 'map[string]interface{}' is expected but '%T' is found", obj)
}

// Encode Object
func (d *Object) Encode() (interface{}, error) {
	obj, err := d.object.Encode()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Decode Object
func (d *Object) Decode(obj interface{}) (interface{}, error) {
	if value, ok := obj.(map[string]interface{}); ok {
		if err := d.object.Decode(value); err != nil {
			return nil, err
		}

		d.Base.MarkDefined()
		return d.object, nil
	}

	return nil,
		fmt.Errorf("Object: 'map[string]interface{}' is expected but '%T' is found", obj)
}

// ToJSON prepares the data for MarshalJSON
func (d *Object) ToJSON() (interface{}, error) {
	return d.object.GetData()
}

// Resolve resolves the value
func (d *Object) Resolve(path []string) (interface{}, []string, error) {
	return d.object.Resolve(path)
}
