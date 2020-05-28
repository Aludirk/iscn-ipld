package data

import (
	"fmt"

	"github.com/ipfs/go-cid"

	node "github.com/ipfs/go-ipld-format"
)

// ==================================================
// Cid
// ==================================================

// Cid is a data handler for IPFS CID
type Cid struct {
	*Base

	codec uint64
	c     []byte
}

var _ Data = (*Cid)(nil)

// NewCid creates a IPFS CID data handler
func NewCid(key string, isRequired bool, codec uint64) *Cid {
	return &Cid{
		Base:  NewBase(key, isRequired),
		codec: codec,
	}
}

// Prototype creates a prototype Cid
func (d *Cid) Prototype() Data {
	return &Cid{
		Base:  d.Base.Prototype(),
		codec: d.codec,
	}
}

// Link returns a link object for IPLD
func (d *Cid) Link() (*node.Link, error) {
	_, c, err := cid.CidFromBytes(d.c)
	if err != nil {
		return nil, err
	}

	return &node.Link{Cid: c}, nil
}

// Set the value of IPFS CID
func (d *Cid) Set(obj interface{}) error {
	if c, ok := obj.(cid.Cid); ok {
		if d.codec != 0 && c.Type() != d.codec {
			return fmt.Errorf(
				"Cid: Codec '0x%x' is expected but '0x%x' is found",
				d.codec,
				c.Type())
		}

		d.c = c.Bytes()
		d.Base.MarkDefined()
		return nil
	}

	return fmt.Errorf("Cid: 'cid.Cid' is expected but '%T' is found", obj)
}

// Encode Cid
func (d *Cid) Encode() (interface{}, error) {
	return d.c, nil
}

// Decode Cid
func (d *Cid) Decode(obj interface{}) (interface{}, error) {
	c, ok := obj.([]byte)
	if !ok {
		return nil,
			fmt.Errorf("Unknown error during decoding Cid: "+
				"'[]byte' is expected but '%T' is found",
				obj,
			)
	}

	_, value, err := cid.CidFromBytes(c)
	if err != nil {
		return nil, err
	}

	if d.codec != 0 && value.Type() != d.codec {
		return nil,
			fmt.Errorf(
				"Cid: Codec '0x%x' is expected but '0x%x' is found",
				d.codec,
				value.Type(),
			)
	}

	d.c = c
	d.Base.MarkDefined()
	return value, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *Cid) ToJSON() (interface{}, error) {
	_, c, err := cid.CidFromBytes(d.c)
	if err != nil {
		return nil, err
	}

	value, err := c.StringOfBase('z')
	if err != nil {
		return nil, err
	}

	link := map[string]string{
		"/": fmt.Sprintf("/ipfs/%s", value),
	}

	return link, nil
}

// Resolve resolves the link
func (d *Cid) Resolve(path []string) (interface{}, []string, error) {
	link, err := d.Link()
	if err != nil {
		return nil, nil, err
	}

	return link, path, nil
}
