package stakeholder

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
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
	*data.FilterString
}

var _ data.Data = (*Type)(nil)

// NewType creates a stakeholder type data handler
func NewType() *Type {
	return &Type{
		FilterString: data.NewFilterString(
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
func (d *Type) Prototype() data.Data {
	return NewType()
}

// ==================================================
// Footprint
// ==================================================

// Footprint is a data handler for the footprint link to the underlying work
type Footprint struct {
	*data.Base

	handler data.Data
}

var _ data.Data = (*Footprint)(nil)

// NewFootprint creates a footprint data handler
func NewFootprint() *Footprint {
	return &Footprint{
		Base: data.NewBase("footprint", false),
	}
}

// Prototype creates a protype Footprint
func (d *Footprint) Prototype() data.Data {
	return &Footprint{
		Base: d.Base.Prototype(),
	}
}

// Set the value of link of footprint
func (d *Footprint) Set(obj interface{}) error {
	if d.handler != nil {
		return fmt.Errorf("Footprint: re-create handler")
	}

	switch obj.(type) {
	case cid.Cid:
		d.handler = data.NewCid(d.GetKey(), d.IsRequired(), block.CodecISCN)
	case string:
		// TODO URL handler
		d.handler = data.NewString(d.GetKey(), d.IsRequired())
	default:
		return fmt.Errorf("Footprint: link is expected but '%T' is found", obj)
	}

	if err := d.handler.Set(obj); err != nil {
		return err
	}

	d.Base.MarkDefined()
	return nil
}

// Encode Footprint
func (d *Footprint) Encode() (interface{}, error) {
	return d.handler.Encode()
}

// Decode Footprint
func (d *Footprint) Decode(obj interface{}) (interface{}, error) {
	if d.handler != nil {
		return nil, fmt.Errorf("Footprint: re-create handler")
	}

	switch obj.(type) {
	case []uint8:
		d.handler = data.NewCid(d.GetKey(), d.IsRequired(), block.CodecISCN)
	case string:
		// TODO URL handler
		d.handler = data.NewString(d.GetKey(), d.IsRequired())
	default:
		return nil, fmt.Errorf("Footprint: link is expected but '%T' is found", obj)
	}

	dec, err := d.handler.Decode(obj)
	if err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
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
