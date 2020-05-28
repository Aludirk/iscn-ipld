package block

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ipfs/go-cid"
	"github.com/likecoin/iscn-ipld/plugin/block/data"
	"gitlab.com/c0b/go-ordered-json"

	blocks "github.com/ipfs/go-block-format"
	cbor "github.com/ipfs/go-ipld-cbor"
	node "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"
)

// ==================================================
// IscnObject
// ==================================================

// IscnObject is the interface for the data object of ISCN object
type IscnObject interface {
	node.Node

	GetName() string
	GetVersion() uint64
	GetCustom() map[string]interface{}

	GetArray(string) ([]interface{}, error)
	GetObject(string) (interface{}, error)
	GetBytes(string) ([]byte, error)
	GetInt32(string) (int32, error)
	GetUint32(string) (uint32, error)
	GetInt64(string) (int64, error)
	GetUint64(string) (uint64, error)
	GetString(string) (string, error)
	GetCid(string) (cid.Cid, error)
	GetLink(string) (cid.Cid, string, error)

	MarshalJSON() ([]byte, error)
}

// ==================================================
// Codec
// ==================================================

// Codec is the interface for the CODEC of ISCN object
type Codec interface {
	IscnObject

	MarkNested()

	GetData() (*ordered.OrderedMap, error)
	SetData(map[string]interface{}) error

	Encode() (map[string]interface{}, error)
	Decode(map[string]interface{}) error
}

// CodecFactoryFunc returns a factory function to create ISCN object
type CodecFactoryFunc func() (Codec, error)

type codecFactory map[uint64][]CodecFactoryFunc

var factory codecFactory = codecFactory{}
var schemaNames map[uint64]string = map[uint64]string{}

// Validator is a validate function for post validation after set data in block
type Validator func() error

// RegisterIscnObjectFactory registers an array of ISCN object factory functions
func RegisterIscnObjectFactory(
	codec uint64,
	schemaName string,
	factories ...CodecFactoryFunc,
) {
	factory[codec] = factories
	schemaNames[codec] = schemaName
}

// Encode the data to specific ISCN object and version
func Encode(
	codec uint64,
	version uint64,
	m map[string]interface{},
) (IscnObject, error) {
	schemas, ok := factory[codec]
	if !ok {
		return nil, fmt.Errorf("Codec 0x%x is not registered", codec)
	}

	if version > (uint64)(len(schemas)) {
		return nil, fmt.Errorf("<%s (v%d)> is not implemented", schemaNames[codec], version)
	}
	version--

	obj, err := schemas[version]()
	if err != nil {
		return nil, err
	}

	if err := obj.SetData(m); err != nil {
		return nil, err
	}

	if _, err := obj.Encode(); err != nil {
		return nil, err
	}

	return obj, nil
}

// DecodeBlock decodes the raw IPLD data back to data object
func DecodeBlock(block blocks.Block) (node.Node, error) {
	return Decode(block.RawData(), block.Cid())
}

// Decode decodes the raw IPLD data back to data object
func Decode(rawData []byte, c cid.Cid) (IscnObject, error) {
	rawObj := map[string]interface{}{}
	if err := cbor.DecodeInto(rawData, &rawObj); err != nil {
		return nil, err
	}

	v, ok := rawObj[data.ContextKey]
	if !ok {
		return nil, fmt.Errorf("Invalid ISCN IPLD object, missing context")
	}

	version, ok := v.(uint64)
	if !ok {
		return nil, fmt.Errorf("Context: 'uint64' is expected but '%T' is found", v)
	}

	schemas, ok := factory[c.Type()]
	if !ok {
		return nil, fmt.Errorf("Codec 0x%x is not registered", c.Type())
	}

	if version > (uint64)(len(schemas)) {
		return nil, fmt.Errorf("<%s (v%d)> is not implemented", schemaNames[c.Type()], version)
	}
	version--

	obj, err := schemas[version]()
	if err != nil {
		return nil, err
	}

	if err := obj.Decode(rawObj); err != nil {
		return nil, err
	}

	// Encode one more time to retrieve CID
	if _, err := obj.Encode(); err != nil {
		return nil, err
	}

	// Verify the CID
	if !obj.Cid().Equals(c) {
		current, err := obj.Cid().StringOfBase('z')
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve current CID")
		}

		expected, err := c.StringOfBase('z')
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve expected CID")
		}

		return nil, fmt.Errorf("Cid %q is not matched: expected %q", current, expected)
	}

	return obj, nil
}

const (
	// TODO real domain
	domainIscn = "iscn"

	// TODO real domain
	domainLikeCoin = "likecoin"
)

func getSchema(codec uint64) string {
	switch codec {
	case CodecISCN,
		CodecRights,
		CodecStakeholders,
		CodecContent,
		CodecEntity,
		CodecRight,
		CodecStakeholder,
		CodecTimePeriod:
		return fmt.Sprintf("https://%s/%s", domainIscn, schemaNames[codec])
	}

	panic(fmt.Sprintf("Unknown codec 0x%x", codec))
}

// ==================================================
// Base
// ==================================================

// Base is the basic block of all kind of ISCN objects
type Base struct {
	isNested bool

	codec     uint64
	name      string
	version   uint64
	obj       map[string]interface{}
	data      map[string]data.Data
	keys      []string
	custom    map[string]interface{}
	validator Validator

	cid     *cid.Cid
	rawData []byte
}

var _ Codec = (*Base)(nil)
var _ data.Codec = (*Base)(nil)

// NewBase creates the basic block of an ISCN object
func NewBase(codec uint64, name string, version uint64, schema []data.Data) (*Base, error) {
	// Create the base
	b := &Base{
		isNested:  false,
		codec:     codec,
		name:      name,
		version:   version,
		data:      map[string]data.Data{},
		keys:      []string{},
		custom:    map[string]interface{}{},
		validator: nil,
	}

	// Set "context" data
	context := data.NewContext(getSchema(codec))
	err := context.Set(version)
	if err != nil {
		return nil, err
	}

	b.data[context.GetKey()] = context
	b.keys = append(b.keys, context.GetKey())

	// Setup schema
	for _, handler := range schema {
		key := handler.GetKey()
		b.keys = append(b.keys, key)
		b.data[key] = handler
	}

	return b, nil
}

// GetName returns the name of the the ISCN object
func (b *Base) GetName() string {
	return b.name
}

// GetVersion returns the schema version of the ISCN object
func (b *Base) GetVersion() uint64 {
	return b.version
}

// GetCustom returns the custom data
func (b *Base) GetCustom() map[string]interface{} {
	return b.custom
}

// GetArray returns the value of 'key' as array
func (b *Base) GetArray(key string) ([]interface{}, error) {
	value, ok := b.obj[key]
	if !ok {
		return nil, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("The value of %q is not '[]interface{]'", key)
	}

	return res, nil
}

// GetObject returns the value of 'key' as object
func (b *Base) GetObject(key string) (interface{}, error) {
	value, ok := b.obj[key]
	if !ok {
		return nil, fmt.Errorf("%q is not found", key)
	}

	return value, nil
}

// GetBytes returns the value of 'key' as byte slice
func (b *Base) GetBytes(key string) ([]byte, error) {
	value, ok := b.obj[key]
	if !ok {
		return nil, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.([]byte)
	if !ok {
		return nil, fmt.Errorf("The value of %q is not '[]byte'", key)
	}

	return res, nil
}

// GetInt32 returns the value of 'key' as int32
func (b *Base) GetInt32(key string) (int32, error) {
	value, ok := b.obj[key]
	if !ok {
		return 0, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(int32)
	if !ok {
		return 0, fmt.Errorf("The value of %q is not 'int32'", key)
	}

	return res, nil
}

// GetUint32 returns the value of 'key' as int32
func (b *Base) GetUint32(key string) (uint32, error) {
	value, ok := b.obj[key]
	if !ok {
		return 0, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(uint32)
	if !ok {
		return 0, fmt.Errorf("The value of %q is not 'uint32'", key)
	}

	return res, nil
}

// GetInt64 returns the value of 'key' as int64
func (b *Base) GetInt64(key string) (int64, error) {
	value, ok := b.obj[key]
	if !ok {
		return 0, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("The value of %q is not 'int64'", key)
	}

	return res, nil
}

// GetUint64 returns the value of 'key' as int64
func (b *Base) GetUint64(key string) (uint64, error) {
	value, ok := b.obj[key]
	if !ok {
		return 0, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(uint64)
	if !ok {
		return 0, fmt.Errorf("The value of %q is not 'uint64'", key)
	}

	return res, nil
}

// GetString returns the value of 'key' as string
func (b *Base) GetString(key string) (string, error) {
	value, ok := b.obj[key]
	if !ok {
		return "", fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("The value of %q is not '[]byte'", key)
	}

	return res, nil
}

// GetCid returns the value of 'key' as Cid
func (b *Base) GetCid(key string) (cid.Cid, error) {
	value, ok := b.obj[key]
	if !ok {
		return cid.Undef, fmt.Errorf("%q is not found", key)
	}

	res, ok := value.(cid.Cid)
	if !ok {
		return cid.Undef, fmt.Errorf("The value of %q is not 'Cid'", key)
	}

	return res, nil
}

// GetLink returns the value of 'key' as link, a link can be a Cid or a URL
func (b *Base) GetLink(key string) (cid.Cid, string, error) {
	value, ok := b.obj[key]
	if !ok {
		return cid.Undef, "", fmt.Errorf("%q is not found", key)
	}

	switch v := value.(type) {
	case cid.Cid:
		return v, "", nil
	case string:
		return cid.Undef, v, nil
	}

	return cid.Undef, "", fmt.Errorf("The value of %q is not a link", key)
}

// SetValidator sets the validator function
func (b *Base) SetValidator(validator Validator) {
	b.validator = validator
}

// MarshalJSON convert the block to JSON format
func (b *Base) MarshalJSON() ([]byte, error) {
	om, err := b.GetData()
	if err != nil {
		return nil, err
	}

	return om.MarshalJSON()
}

// MarkNested marks the block as a nested block
func (b *Base) MarkNested() {
	b.isNested = true
}

// GetData returns the block data as ordered.OrderedMap
func (b *Base) GetData() (*ordered.OrderedMap, error) {
	om := ordered.NewOrderedMap()
	for _, key := range b.keys {
		handler := b.data[key]

		if key != data.ContextKey { // Context key does not exist in b.obj
			_, exist := b.obj[key]
			if !exist {
				if handler.IsRequired() {
					return nil, fmt.Errorf("Unknown error: key %q should be exist", key)
				}
				continue
			}
		} else if b.isNested {
			// Nested block do not marshal context
			continue
		}

		value, err := handler.ToJSON()
		if err != nil {
			return nil, err
		}

		om.Set(handler.GetKey(), value)
	}

	for key, value := range b.custom {
		om.Set(key, value)
	}

	return om, nil
}

// SetData sets and validates the data
func (b *Base) SetData(m map[string]interface{}) error {
	b.obj = map[string]interface{}{}

	// Set the data
	for key, handler := range b.data {
		// Skip context property
		if key == data.ContextKey {
			continue
		}

		d, ok := m[key]
		if !ok || d == nil {
			if handler.IsRequired() {
				return fmt.Errorf("The property %q is required", key)
			}

			continue
		}

		err := handler.Set(d)
		if err != nil {
			return err
		}

		// Save the data object
		b.obj[key] = d
	}

	// Validate the data
	if b.validator != nil {
		if err := b.validator(); err != nil {
			return err
		}
	}

	// Save the custom data
	for key, value := range m {
		_, exist := b.data[key]
		if !exist {
			b.custom[key] = value
		}
	}

	return nil
}

// Encode the ISCN object to CBOR serialized data
func (b *Base) Encode() (map[string]interface{}, error) {
	// Extract all data from data handlers
	m := map[string]interface{}{}
	for _, handler := range b.data {
		if handler.GetKey() != data.ContextKey { // Context key does not exist in b.obj
			_, exist := b.obj[handler.GetKey()]
			if !exist {
				if handler.IsRequired() {
					return nil,
						fmt.Errorf("Unknown error: key %q should be exist", handler.GetKey())
				}
				continue
			}
		} else if b.isNested {
			// Nested block do not encode context
			continue
		}

		enc, err := handler.Encode()
		if err != nil {
			return nil, err
		}
		m[handler.GetKey()] = enc
	}

	// Merge the custom data
	for key, value := range b.custom {
		m[key] = value
	}

	// CBOR-ise the data
	rawData, err := cbor.DumpObject(m)
	if err != nil {
		return nil, err
	}

	c, err := cid.V1Builder{
		Codec:  b.codec,
		MhType: mh.SHA2_256,
	}.Sum(rawData)
	if err != nil {
		return nil, err
	}

	b.cid = &c
	b.rawData = rawData

	return m, nil
}

// Decode the data back to ISCN object
func (b *Base) Decode(m map[string]interface{}) error {
	// Remove the context property as it is processed by base ISCN object
	delete(m, data.ContextKey)

	b.obj = map[string]interface{}{}
	for key, handler := range b.data {
		// Skip context property
		if key == data.ContextKey {
			continue
		}

		d, ok := m[key]
		if !ok || d == nil {
			if handler.IsRequired() {
				return fmt.Errorf("The property %q is required", key)
			}

			continue
		}

		dec, err := handler.Decode(d)
		if err != nil {
			return err
		}
		b.obj[key] = dec

		delete(m, key)
	}

	// Validate the data
	if b.validator != nil {
		if err := b.validator(); err != nil {
			return err
		}
	}

	// Save the custom data
	b.custom = m

	return nil
}

// github.com/ipfs/go-block-format.Block interface

// Cid returns the CID of the block header
func (b *Base) Cid() cid.Cid {
	return *(b.cid)
}

// Loggable returns a map the type of IPLD Link
func (b *Base) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"type":    b.name,
		"version": b.version,
	}
}

// RawData returns the binary of the CBOR encode of the block header
func (b *Base) RawData() []byte {
	return b.rawData
}

// String is a helper for output
func (b *Base) String() string {
	return fmt.Sprintf("<%s (v%d)>", b.GetName(), b.GetVersion())
}

// node.Resolver interface

// Resolve resolves a path through this node, stopping at any link boundary
// and returning the object found as well as the remaining path to traverse
func (b *Base) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		return b, nil, nil
	}

	first, rest := path[0], path[1:]

	if handler, ok := b.data[first]; ok {
		if first != data.ContextKey { // Context key does not exist in b.obj
			_, exist := b.obj[first]
			if !exist {
				if handler.IsRequired() {
					return nil, nil, fmt.Errorf("Unknown error: key %q should be exist", first)
				}
				return nil, nil, fmt.Errorf("no such link")
			}
		} else if b.isNested {
			// Nested block do not resolve context
			return nil, nil, fmt.Errorf("no such link")
		}
		return handler.Resolve(rest)
	}

	// Handle custom parameters
	var obj interface{} = b.custom
	for _, key := range path {
		switch value := obj.(type) {
		case map[string]interface{}:
			v, ok := value[key]
			if !ok {
				return nil, nil, fmt.Errorf("no such link")
			}
			obj = v
		case []interface{}:
			i, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("no such link")
			}

			if i >= int64(len(value)) {
				return nil, nil, fmt.Errorf("index %d does not exist", i)
			}

			obj = value[i]
		default:
			return nil, nil, fmt.Errorf("no such link")
		}
	}

	return obj, nil, nil
}

// Tree lists all paths within the object under 'path', and up to the given depth.
// To list the entire object (similar to `find .`) pass "" and -1
func (*Base) Tree(path string, depth int) []string {
	log.Println("Tree is not implemented")
	return nil
}

// node.Node interface

// Copy will go away. It is here to comply with the Node interface.
func (*Base) Copy() node.Node {
	panic("dont use this yet")
}

// Links is a helper function that returns all links within this object
// HINT: Use `ipfs refs <cid>`
func (b *Base) Links() []*node.Link {
	links := []*node.Link{}
	for _, d := range b.data {
		switch c := d.(type) {
		case *data.Cid:
			if link, err := c.Link(); err == nil {
				links = append(links, link)
			}
		}
	}
	return links
}

// ResolveLink is a helper function that allows easier traversal of links through blocks
func (b *Base) ResolveLink(path []string) (*node.Link, []string, error) {
	obj, rest, err := b.Resolve(path)
	if err != nil {
		return nil, nil, err
	}

	if lnk, ok := obj.(*node.Link); ok {
		return lnk, rest, nil
	}

	return nil, nil, fmt.Errorf("resolved item was not a link")
}

// Size will go away. It is here to comply with the Node interface.
func (*Base) Size() (uint64, error) {
	return 0, nil
}

// Stat will go away. It is here to comply with the Node interface.
func (*Base) Stat() (*node.NodeStat, error) {
	return &node.NodeStat{}, nil
}
