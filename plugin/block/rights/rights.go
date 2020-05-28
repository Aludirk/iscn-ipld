package rights

import (
	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
	"github.com/likecoin/iscn-ipld/plugin/block/right"
)

const (
	// SchemaName of rights
	SchemaName = "rights"
)

// Register registers the schema of rights block
func Register() {
	block.RegisterIscnObjectFactory(
		block.CodecRights,
		SchemaName,
		newSchemaV1,
	)
}

// ==================================================
// base
// ==================================================

// base is the base struct for rights (codec 0x0265)
type base struct {
	*block.Base
}

func newBase(version uint64, schema []data.Data) (*base, error) {
	blockBase, err := block.NewBase(
		block.CodecRights,
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

// schemaV1 represents a rights V1
type schemaV1 struct {
	*base
}

var _ block.IscnObject = (*schemaV1)(nil)

func newSchemaV1() (block.Codec, error) {
	prototype := data.NewObject("_", true, right.SchemaV1Prototype)

	schema := []data.Data{
		data.NewDataArray("rights", true, prototype),
	}

	rightsBase, err := newBase(1, schema)
	if err != nil {
		return nil, err
	}

	return &schemaV1{
		base: rightsBase,
	}, nil
}
