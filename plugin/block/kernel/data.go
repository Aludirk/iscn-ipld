package kernel

import (
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
)

// ==================================================
// ID
// ==================================================

// ID is a data handler for the ISCN ID
type ID struct {
	*data.Base

	id []byte
}

var _ data.Data = (*ID)(nil)

// NewID creates an ISCN ID
func NewID() *ID {
	return &ID{
		Base: data.NewBase("id", true),
	}
}

// Prototype creates a prototype ID
func (d *ID) Prototype() data.Data {
	return &ID{
		Base: d.Base.Prototype(),
	}
}

// GetID returns the human readable ID
func (d *ID) GetID() string {
	return fmt.Sprintf("1/%s", base58.Encode(d.id))
}

// Set the value of ID
func (d *ID) Set(obj interface{}) error {
	if id, ok := obj.([]byte); ok {
		if len(id) != 32 {
			return fmt.Errorf("ID: should length 32 but %d is found", len(id))
		}

		d.id = id
		d.Base.MarkDefined()
		return nil
	}

	return fmt.Errorf("ID: '[]byte' is expected but '%T' is found", obj)
}

// Encode ID
func (d *ID) Encode() (interface{}, error) {
	return d.id, nil
}

// Decode ID
func (d *ID) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.id, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *ID) ToJSON() (interface{}, error) {
	return d.GetID(), nil
}

// Resolve resolves the value
func (d *ID) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.GetID(), nil, nil
}
