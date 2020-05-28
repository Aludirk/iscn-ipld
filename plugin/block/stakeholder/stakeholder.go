package stakeholder

import (
	"fmt"

	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
)

const (
	// SchemaName of stakeholder
	SchemaName = "stakeholder"
)

// Register registers the schema of stakeholder block
func Register() {
	block.RegisterIscnObjectFactory(
		block.CodecStakeholder,
		SchemaName,
		newSchemaV1,
	)
}

// ==================================================
// base
// ==================================================

// base is the base struct for stakeholder (codec 0x02D1)
type base struct {
	*block.Base
}

func newBase(version uint64, schema []data.Data) (*base, error) {
	blockBase, err := block.NewBase(
		block.CodecStakeholder,
		SchemaName,
		version,
		schema,
	)
	if err != nil {
		return nil, err
	}

	return &base{
		Base: blockBase,
	}, nil
}

// ==================================================
// schemaV1
// ==================================================

// schemaV1 represents a stakeholder V1
type schemaV1 struct {
	*base

	typ       *Type
	footprint *Footprint
}

var _ block.IscnObject = (*schemaV1)(nil)

func newSchemaV1() (block.Codec, error) {
	typ := NewType()
	footprint := NewFootprint()

	schema := []data.Data{
		typ,
		data.NewCid("stakeholder", true, block.CodecEntity),
		data.NewNumber("sharing", true, data.Uint32T),
		footprint,
	}

	stakeholderBase, err := newBase(1, schema)
	if err != nil {
		return nil, err
	}

	obj := schemaV1{
		base:      stakeholderBase,
		typ:       typ,
		footprint: footprint,
	}
	stakeholderBase.SetValidator(obj.Validate)

	return &obj, nil
}

// SchemaV1Prototype creates a prototype for schemaV1
func SchemaV1Prototype() data.Codec {
	res, _ := newSchemaV1()
	return res
}

// Validate the data
func (o *schemaV1) Validate() error {
	if o.typ.Get() == footprint {
		if !o.footprint.IsDefined() {
			return fmt.Errorf("Footprint is missed")
		}
	} else {
		if o.footprint.IsDefined() {
			return fmt.Errorf("Footprint should not be set as this is not a footprint stakeholder")
		}
	}

	return nil
}
