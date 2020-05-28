package entity

import (
	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
)

const (
	// SchemaName of entity
	SchemaName = "entity"
)

// Register registers the schema of entity block
func Register() {
	block.RegisterIscnObjectFactory(
		block.CodecEntity,
		SchemaName,
		newSchemaV1,
	)
}

// ==================================================
// base
// ==================================================

// base is the base struct for entity (codec 0x0268)
type base struct {
	*block.Base
}

func newBase(version uint64, schema []data.Data) (*base, error) {
	blockBase, err := block.NewBase(
		block.CodecEntity,
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

// schemaV1 represents an entity V1
type schemaV1 struct {
	*base
}

var _ block.IscnObject = (*schemaV1)(nil)

func newSchemaV1() (block.Codec, error) {
	schema := []data.Data{
		data.NewString("id", true), // TODO llc://id
		data.NewString("name", false),
		data.NewString("description", false),
	}

	entityBase, err := newBase(1, schema)
	if err != nil {
		return nil, err
	}

	return &schemaV1{
		base: entityBase,
	}, nil
}
