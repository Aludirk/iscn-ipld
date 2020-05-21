package iscn

import (
	"github.com/likecoin/iscn-ipld/plugin/block"
	"github.com/likecoin/iscn-ipld/plugin/block/content"
	"github.com/likecoin/iscn-ipld/plugin/block/entity"
	"github.com/likecoin/iscn-ipld/plugin/block/kernel"
	"github.com/likecoin/iscn-ipld/plugin/block/right"
	"github.com/likecoin/iscn-ipld/plugin/block/rights"
	"github.com/likecoin/iscn-ipld/plugin/block/stakeholder"
	"github.com/likecoin/iscn-ipld/plugin/block/stakeholders"
	"github.com/likecoin/iscn-ipld/plugin/block/time_period"

	ipld "github.com/ipfs/go-ipld-format"
)

// Register all ISCN objects
func Register() {
	kernel.Register()
	rights.Register()
	stakeholders.Register()
	content.Register()
	entity.Register()

	right.Register()
	stakeholder.Register()
	timeperiod.Register()
}

// RegisterBlockDecoders registers the decoder for different types of ISCN block
func RegisterBlockDecoders(decoder ipld.BlockDecoder) error {
	decoder.Register(block.CodecISCN, block.DecodeBlock)
	decoder.Register(block.CodecRights, block.DecodeBlock)
	decoder.Register(block.CodecStakeholders, block.DecodeBlock)
	decoder.Register(block.CodecContent, block.DecodeBlock)
	decoder.Register(block.CodecEntity, block.DecodeBlock)
	return nil
}
