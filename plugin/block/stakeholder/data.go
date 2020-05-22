package stakeholder

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/likecoin/iscn-ipld/plugin/block"
)

// ==================================================
// Type
// ==================================================

const (
	creator     = "Creator"
	contributor = "Contributor"
	editor      = "Editor"
	publisher   = "Publisher"
	footprint   = "FootprintStakeholder"
	escrow      = "Escrow"
)

// Type is a data handler for the type of stakeholder
type Type struct {
	*block.String
}

var _ block.Data = (*Type)(nil)

// NewType creates a stakeholder type data handler
func NewType() *Type {
	return &Type{
		String: block.NewStringWithFilter(
			"type",
			true,
			[]string{
				creator,
				contributor,
				editor,
				publisher,
				footprint,
				escrow,
			},
		),
	}
}

// Prototype creates a prototype Type
func (d *Type) Prototype() block.Data {
	return NewType()
}

// ==================================================
// Footprint
// ==================================================

// Footprint is a data handler for the footprint link to the underlying work
type Footprint struct {
	*block.DataBase

	handler block.Data
}

var _ block.Data = (*Footprint)(nil)

// NewFootprint creates a footprint data handler
func NewFootprint() *Footprint {
	return &Footprint{
		DataBase: block.NewDataBase("footprint", false),
	}
}

// Prototype creates a protype Footprint
func (d *Footprint) Prototype() block.Data {
	return &Footprint{
		DataBase: d.DataBase.Prototype(),
	}
}

// Set the value of link of footprint
func (d *Footprint) Set(data interface{}) error {
	if d.handler != nil {
		return fmt.Errorf("Footprint: re-create handler")
	}

	switch data.(type) {
	case cid.Cid:
		d.handler = block.NewCid(d.GetKey(), d.IsRequired(), block.CodecISCN)
	case string:
		// TODO URL handler
		d.handler = block.NewString(d.GetKey(), d.IsRequired())
	default:
		return fmt.Errorf("Footprint: link is expected but '%T' is found", data)
	}

	if err := d.handler.Set(data); err != nil {
		return err
	}

	return d.DataBase.Set(data)
}

// Encode Footprint
func (d *Footprint) Encode() (interface{}, error) {
	return d.handler.Encode()
}

// Decode Footprint
func (d *Footprint) Decode(data interface{}) (interface{}, error) {
	if d.handler != nil {
		return nil, fmt.Errorf("Footprint: re-create handler")
	}

	switch data.(type) {
	case []uint8:
		d.handler = block.NewCid(d.GetKey(), d.IsRequired(), block.CodecISCN)
	case string:
		// TODO URL handler
		d.handler = block.NewString(d.GetKey(), d.IsRequired())
	default:
		return nil, fmt.Errorf("Footprint: link is expected but '%T' is found", data)
	}

	dec, err := d.handler.Decode(data)
	if err != nil {
		return nil, err
	}

	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

	return dec, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *Footprint) ToJSON() (interface{}, error) {
	return d.handler.ToJSON()
}

// Resolve resolves the link
func (d *Footprint) Resolve(path []string) (interface{}, []string, error) {
	return d.handler.Resolve(path)
}
