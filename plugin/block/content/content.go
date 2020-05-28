package content

import (
	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
)

const (
	// SchemaName of content
	SchemaName = "content"
)

// Register registers the schema of content block
func Register() {
	block.RegisterIscnObjectFactory(
		block.CodecContent,
		SchemaName,
		newSchemaV1,
	)
}

// ==================================================
// base
// ==================================================

// base is the base struct for content (codec 0x0267)
type base struct {
	*block.Base
}

func newBase(version uint64, schema []data.Data) (*base, error) {
	blockBase, err := block.NewBase(
		block.CodecContent,
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

// schemaV1 represents a content V1
type schemaV1 struct {
	*base

	version *data.Number
	parent  *data.Cid
}

var _ block.IscnObject = (*schemaV1)(nil)

func newSchemaV1() (block.Codec, error) {
	version := data.NewNumber("version", true, data.Uint64T)
	parent := data.NewCid("parent", false, block.CodecContent)

	schema := []data.Data{
		data.NewString("type", true),
		version,
		parent,
		data.NewURL("source", false),
		data.NewString("edition", false),
		data.NewHash("fingerprint", true),
		data.NewString("title", true),
		data.NewString("description", false),
		data.NewDataArray("tags", false, data.NewString("_", false)),
	}

	contentBase, err := newBase(1, schema)
	if err != nil {
		return nil, err
	}

	obj := schemaV1{
		base:    contentBase,
		version: version,
		parent:  parent,
	}
	contentBase.SetValidator(obj.Validate)

	return &obj, nil
}

// Validate the data
func (o *schemaV1) Validate() error {
	return data.ValidateParent(o.version, o.parent)
}
