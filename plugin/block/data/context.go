package data

import (
	"fmt"
)

// ==================================================
// Context
// ==================================================

const (
	// ContextKey is the key of context of ISCN object
	ContextKey = "context"
)

// Context is a data handler for the context of ISCN object
type Context struct {
	*Base

	handler *Number
	schema  string
	version uint64
}

var _ Data = (*Context)(nil)

// NewContext creates a context of ISCN object with schema
func NewContext(schema string) *Context {
	return &Context{
		Base:    NewBase(ContextKey, true),
		handler: NewNumber("", false, Uint64T),
		schema:  schema,
	}
}

// Prototype creates a prototype Context
func (d *Context) Prototype() Data {
	return &Context{
		Base:    d.Base.Prototype(),
		handler: NewNumber(d.handler.GetKey(), d.handler.IsRequired(), d.handler.typ),
		schema:  d.schema,
	}
}

// Set the value of Context
func (d *Context) Set(obj interface{}) error {
	err := d.handler.Set(obj)
	if err != nil {
		return fmt.Errorf("Context: 'uint64' is expected but '%T' is found", obj)
	}

	version, err := d.handler.GetUint64()
	if err != nil {
		return fmt.Errorf("Context: %s", err)
	}

	d.version = version
	d.Base.MarkDefined()
	return nil
}

// Encode Context
func (d *Context) Encode() (interface{}, error) {
	return d.version, nil
}

// Decode Context
func (d *Context) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.version, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *Context) ToJSON() (interface{}, error) {
	return d.getSchema(), nil
}

// Resolve resolves the value
func (d *Context) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.getSchema(), nil, nil
}

func (d *Context) getSchema() string {
	return fmt.Sprintf("%s-v%d", d.schema, d.version)
}
