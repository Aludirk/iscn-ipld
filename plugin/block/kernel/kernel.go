package kernel

import (
	"fmt"

	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
)

const (
	// SchemaName of ISCN kernel
	SchemaName = "iscn"
)

// Register registers the schema of ISCN kernel block
func Register() {
	block.RegisterIscnObjectFactory(
		block.CodecISCN,
		SchemaName,
		newSchemaV1,
	)
}

// ==================================================
// base
// ==================================================

// base is the base struct for ISCN kernel (codec 0x0264)
type base struct {
	*block.Base

	id *ID
}

func newBase(version uint64, schema []data.Data, id *ID) (*base, error) {
	blockBase, err := block.NewBase(
		block.CodecISCN,
		SchemaName,
		version,
		schema,
	)
	if err != nil {
		return nil, err
	}

	return &base{
		Base: blockBase,
		id:   id,
	}, nil
}

// github.com/ipfs/go-block-format.Block interface

// Loggable returns a map the type of IPLD Link
func (b *base) Loggable() map[string]interface{} {
	l := b.Base.Loggable()
	l["id"] = b.id.GetID()
	return l
}

// String is a helper for output
func (b *base) String() string {
	return fmt.Sprintf("<%s (v%d): %s>", b.GetName(), b.GetVersion(), b.id.GetID())
}

// ==================================================
// schemaV1
// ==================================================

// schemaV1 represents an ISCN kernel V1
type schemaV1 struct {
	*base

	version *data.Number
	parent  *data.Cid
}

var _ block.IscnObject = (*schemaV1)(nil)

func newSchemaV1() (block.Codec, error) {
	id := NewID()
	version := data.NewNumber("version", true, data.Uint64T)
	parent := data.NewCid("parent", false, block.CodecISCN)

	schema := []data.Data{
		id,
		data.NewTimestamp("timestamp", true),
		version,
		parent,
		data.NewCid("rights", true, block.CodecRights),
		data.NewCid("stakeholders", true, block.CodecStakeholders),
		data.NewCid("content", true, block.CodecContent),
	}

	iscnKernelBase, err := newBase(1, schema, id)
	if err != nil {
		return nil, err
	}

	obj := schemaV1{
		base:    iscnKernelBase,
		version: version,
		parent:  parent,
	}
	iscnKernelBase.SetValidator(obj.Validate)

	return &obj, nil
}

// Validate the data
func (o *schemaV1) Validate() error {
	return data.ValidateParent(o.version, o.parent)
}
