package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"

	"github.com/ipfs/go-cid"

	node "github.com/ipfs/go-ipld-format"
)

// ==================================================
// Data
// ==================================================

// Data is the interface for the data property handler
type Data interface {
	Prototype() Data

	IsRequired() bool
	IsDefined() bool

	Set(interface{}) error
	GetKey() string

	Encode() (interface{}, error)
	Decode(interface{}) (interface{}, error)
	ToJSON() (interface{}, error)

	Resolve(path []string) (interface{}, []string, error)
}

// ==================================================
// DataBase
// ==================================================

// DataBase is the base struct for handling data property
type DataBase struct {
	isRequired bool
	isDefinded bool

	key string
}

// NewDataBase creates a base struct for handling data property
func NewDataBase(key string, isRequired bool) *DataBase {
	return &DataBase{
		isRequired: isRequired,
		isDefinded: false,
		key:        key,
	}
}

// Prototype creates a prototype DataBase
func (b *DataBase) Prototype() *DataBase {
	return &DataBase{
		isRequired: b.isRequired,
		key:        b.key,
	}
}

// IsRequired checks whether the data handler is required
func (b *DataBase) IsRequired() bool {
	return b.isRequired
}

// IsDefined checks whether the data is well defined
func (b *DataBase) IsDefined() bool {
	return b.isDefinded
}

// GetKey returns the key of the data property
func (b *DataBase) GetKey() string {
	return b.key
}

// Set the value of data
func (b *DataBase) Set(interface{}) error {
	b.isDefinded = true
	return nil
}

// Decode the data
func (b *DataBase) Decode(interface{}) (interface{}, error) {
	b.isDefinded = true
	return nil, nil
}

// ==================================================
// DataArray
// ==================================================

// DataArray is an array of data handler
type DataArray struct {
	*DataBase

	array     []Data
	prototype Data
}

var _ Data = (*DataArray)(nil)

// NewDataArray creates an array of data handler
func NewDataArray(key string, isRequired bool, prototype Data) *DataArray {
	return &DataArray{
		DataBase:  NewDataBase(key, isRequired),
		array:     []Data{},
		prototype: prototype,
	}
}

// Prototype creates a prototype DataArray
func (d *DataArray) Prototype() Data {
	return &DataArray{
		DataBase:  d.DataBase.Prototype(),
		array:     []Data{},
		prototype: d.prototype.Prototype(),
	}
}

// Set the value of data handler array
func (d *DataArray) Set(data interface{}) error {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(data)
		for i := 0; i < s.Len(); i++ {
			elem := d.prototype.Prototype()
			if err := elem.Set(s.Index(i).Interface()); err != nil {
				return fmt.Errorf("(Index %d) %s", i, err.Error())
			}
			d.array = append(d.array, elem)
		}

		return d.DataBase.Set(data)
	}

	return fmt.Errorf("DataArray: an array is expected but '%T' is found", data)
}

// Encode DataArray
func (d *DataArray) Encode() (interface{}, error) {
	res := []interface{}{}
	for i, data := range d.array {
		enc, err := data.Encode()
		if err != nil {
			return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
		}
		res = append(res, enc)
	}

	return res, nil
}

// Decode DataArray
func (d *DataArray) Decode(data interface{}) (interface{}, error) {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		res := []interface{}{}
		s := reflect.ValueOf(data)
		for i := 0; i < s.Len(); i++ {
			elem := d.prototype.Prototype()
			dec, err := elem.Decode(s.Index(i).Interface())
			if err != nil {
				return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
			}

			res = append(res, dec)
			d.array = append(d.array, elem)
		}

		if _, err := d.DataBase.Decode(data); err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, fmt.Errorf("DataArray: an array is expected but '%T' is found", data)
}

// ToJSON prepares the data for MarshalJSON
func (d *DataArray) ToJSON() (interface{}, error) {
	res := []interface{}{}
	for i, data := range d.array {
		value, err := data.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("(Index %d) %s", i, err.Error())
		}
		res = append(res, value)
	}

	return res, nil
}

// Resolve resolves the value
func (d *DataArray) Resolve(path []string) (interface{}, []string, error) {
	if len(path) == 0 {
		res := []interface{}{}
		for i, data := range d.array {
			value, _, err := data.Resolve(path)
			if err != nil {
				return nil, nil, fmt.Errorf("(Index %d) %s", i, err.Error())
			}
			res = append(res, value)
		}
		return res, nil, nil
	}

	first, rest := path[0], path[1:]
	index, err := strconv.ParseUint(first, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	if index >= uint64(len(d.array)) {
		return nil, nil, fmt.Errorf("index %d does not exist", index)
	}

	return d.array[index].Resolve(rest)
}

// ==================================================
// Object
// ==================================================

// ObjectPrototypeFunc returns a factory function to create ISCN object prototype
type ObjectPrototypeFunc func() Codec

// Object is a data handler of a nested ISCN object
type Object struct {
	*DataBase

	prototypeFunc ObjectPrototypeFunc
	object        Codec
}

var _ Data = (*Object)(nil)

// NewObject creates a nested ISCN object data handler
func NewObject(key string, isRequired bool, prototypeFunc ObjectPrototypeFunc) *Object {
	object := prototypeFunc()
	object.MarkNested()

	return &Object{
		DataBase:      NewDataBase(key, isRequired),
		prototypeFunc: prototypeFunc,
		object:        object,
	}
}

// Prototype creates a prototype Object
func (d *Object) Prototype() Data {
	object := d.prototypeFunc()
	object.MarkNested()

	return &Object{
		DataBase:      d.DataBase.Prototype(),
		prototypeFunc: d.prototypeFunc,
		object:        object,
	}
}

// Set the value of Object
func (d *Object) Set(data interface{}) error {
	if value, ok := data.(map[string]interface{}); ok {
		if err := d.object.SetData(value); err != nil {
			return err
		}

		return d.DataBase.Set(data)
	}

	return fmt.Errorf("Object: 'map[string]interface{}' is expected but '%T' is found", data)
}

// Encode Object
func (d *Object) Encode() (interface{}, error) {
	obj, err := d.object.Encode()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Decode Object
func (d *Object) Decode(data interface{}) (interface{}, error) {
	if value, ok := data.(map[string]interface{}); ok {
		if err := d.object.Decode(value); err != nil {
			return nil, err
		}

		if _, err := d.DataBase.Decode(data); err != nil {
			return nil, err
		}

		return d.object, nil
	}

	return nil,
		fmt.Errorf("Object: 'map[string]interface{}' is expected but '%T' is found", data)
}

// ToJSON prepares the data for MarshalJSON
func (d *Object) ToJSON() (interface{}, error) {
	return d.object.GetData()
}

// Resolve resolves the value
func (d *Object) Resolve(path []string) (interface{}, []string, error) {
	return d.object.Resolve(path)
}

// ==================================================
// Number
// ==================================================

// NumberType is a enum type for the type of a number
type NumberType int

const (
	// Int32T represents int32
	Int32T NumberType = iota

	// Uint32T represents uint32
	Uint32T

	// Int64T represents int64
	Int64T

	// Uint64T represents uint64
	Uint64T
)

// Number is a data handler for the number
type Number struct {
	*DataBase

	number []byte
	typ    NumberType

	i32 int32
	u32 uint32
	i64 int64
	u64 uint64
}

var _ Data = (*Number)(nil)

// NewNumber creates a number data handler
func NewNumber(key string, isRequired bool, typ NumberType) *Number {
	return &Number{
		DataBase: NewDataBase(key, isRequired),
		typ:      typ,
	}
}

// Prototype creates a prototype Number
func (d *Number) Prototype() Data {
	return &Number{
		DataBase: d.DataBase.Prototype(),
		typ:      d.typ,
	}
}

// GetType returns the type of the number
func (d *Number) GetType() NumberType {
	return d.typ
}

// GetInt32 returns an int32 value
func (d *Number) GetInt32() (int32, error) {
	if d.GetType() != Int32T {
		return 0, fmt.Errorf("Number: is not 'int32'")
	}
	return d.i32, nil
}

// GetUint32 returns an uint32 value
func (d *Number) GetUint32() (uint32, error) {
	if d.GetType() != Uint32T {
		return 0, fmt.Errorf("Number: is not 'uint32'")
	}
	return d.u32, nil
}

// GetInt64 returns an int64 value
func (d *Number) GetInt64() (int64, error) {
	if d.GetType() != Int64T {
		return 0, fmt.Errorf("Number: is not 'int64'")
	}
	return d.i64, nil
}

// GetUint64 returns an uint64 value
func (d *Number) GetUint64() (uint64, error) {
	if d.GetType() != Uint64T {
		return 0, fmt.Errorf("Number: is not 'uint64'")
	}
	return d.u64, nil
}

// Set the value of number
func (d *Number) Set(data interface{}) error {
	switch d.GetType() {
	case Int32T:
		var value int32
		switch v := data.(type) {
		case int:
			if v < math.MinInt32 || math.MaxInt32 < v {
				return fmt.Errorf("Number: 'int32' is expected but 'int' is found")
			}
			value = int32(v)
		case int8:
			value = int32(v)
		case int16:
			value = int32(v)
		case int32:
			value = v
		case int64:
			if v < math.MinInt32 || math.MaxInt32 < v {
				return fmt.Errorf("Number: 'int32' is expected but 'int64' is found")
			}
			value = int32(v)
		case uint:
			if v > math.MaxInt32 {
				return fmt.Errorf("Number: 'int32' is expected but 'uint' is found")
			}
			value = int32(v)
		case uint8:
			value = int32(v)
		case uint16:
			value = int32(v)
		case uint32:
			if v > math.MaxInt32 {
				return fmt.Errorf("Number: 'int32' is expected but 'uint32' is found")
			}
			value = int32(v)
		case uint64:
			if v > math.MaxInt32 {
				return fmt.Errorf("Number: 'int32' is expected but 'uint64' is found")
			}
			value = int32(v)
		default:
			return fmt.Errorf("Number: 'int32' is expected but '%T' is found", data)
		}

		buffer := make([]byte, binary.MaxVarintLen32)
		n := binary.PutVarint(buffer, int64(value))
		d.number = buffer[:n]
		d.i32 = value
	case Uint32T:
		var value uint32
		switch v := data.(type) {
		case int:
			if v < 0 || math.MaxUint32 < v {
				return fmt.Errorf("Number: 'uint32' is expected but 'int' is found")
			}
			value = uint32(v)
		case int8:
			if v < 0 {
				return fmt.Errorf("Number: 'uint32' is expected but 'int8' is found")
			}
			value = uint32(v)
		case int16:
			if v < 0 {
				return fmt.Errorf("Number: 'uint32' is expected but 'int16' is found")
			}
			value = uint32(v)
		case int32:
			if v < 0 {
				return fmt.Errorf("Number: 'uint32' is expected but 'int32' is found")
			}
			value = uint32(v)
		case int64:
			if v < 0 || math.MaxUint32 < v {
				return fmt.Errorf("Number: 'uint32' is expected but 'int64' is found")
			}
			value = uint32(v)
		case uint:
			if v > math.MaxUint32 {
				return fmt.Errorf("Number: 'uint32' is expected but 'uint' is found")
			}
			value = uint32(v)
		case uint8:
			value = uint32(v)
		case uint16:
			value = uint32(v)
		case uint32:
			value = v
		case uint64:
			if v > math.MaxUint32 {
				return fmt.Errorf("Number: 'uint32' is expected but 'uint64' is found")
			}
			value = uint32(v)
		default:
			return fmt.Errorf("Number: 'uint32' is expected but '%T' is found", data)
		}

		buffer := make([]byte, binary.MaxVarintLen32)
		n := binary.PutUvarint(buffer, uint64(value))
		d.number = buffer[:n]
		d.u32 = value
	case Int64T:
		var value int64
		switch v := data.(type) {
		case int:
			value = int64(v)
		case int8:
			value = int64(v)
		case int16:
			value = int64(v)
		case int32:
			value = int64(v)
		case int64:
			value = v
		case uint:
			if v > math.MaxInt64 {
				return fmt.Errorf("Number: 'int64' is expected but 'uint' is found")
			}
			value = int64(v)
		case uint8:
			value = int64(v)
		case uint16:
			value = int64(v)
		case uint32:
			value = int64(v)
		case uint64:
			if v > math.MaxInt64 {
				return fmt.Errorf("Number: 'int64' is expected but 'uint64' is found")
			}
			value = int64(v)
		default:
			return fmt.Errorf("Number: 'int64' is expected but '%T' is found", data)
		}

		buffer := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buffer, value)
		d.number = buffer[:n]
		d.i64 = value
	case Uint64T:
		var value uint64
		switch v := data.(type) {
		case int:
			if v < 0 {
				return fmt.Errorf("Number: 'uint64' is expected but 'int' is found")
			}
			value = uint64(v)
		case int8:
			if v < 0 {
				return fmt.Errorf("Number: 'uint64' is expected but 'int8' is found")
			}
			value = uint64(v)
		case int16:
			if v < 0 {
				return fmt.Errorf("Number: 'uint64' is expected but 'int16' is found")
			}
			value = uint64(v)
		case int32:
			if v < 0 {
				return fmt.Errorf("Number: 'uint64' is expected but 'int32' is found")
			}
			value = uint64(v)
		case int64:
			if v < 0 {
				return fmt.Errorf("Number: 'uint64' is expected but 'int64' is found")
			}
			value = uint64(v)
		case uint:
			value = uint64(v)
		case uint8:
			value = uint64(v)
		case uint16:
			value = uint64(v)
		case uint32:
			value = uint64(v)
		case uint64:
			value = v
		default:
			return fmt.Errorf("Number: 'uint64' is expected but '%T' is found", data)
		}

		buffer := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buffer, value)
		d.number = buffer[:n]
		d.u64 = value
	}

	return d.DataBase.Set(data)
}

// Encode Number
func (d *Number) Encode() (interface{}, error) {
	return d.number, nil
}

// Decode Number
func (d *Number) Decode(data interface{}) (interface{}, error) {
	number, ok := data.([]byte)
	if !ok {
		return nil,
			fmt.Errorf("Unknown error during decoding number: "+
				"'[]byte' is expected but '%T' is found",
				data,
			)
	}

	r := bytes.NewReader(number)
	switch d.GetType() {
	case Int32T:
		value, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}

		if value < math.MinInt32 || math.MaxInt32 < value {
			return nil, fmt.Errorf("Unknown error: the number is not an int32")
		}

		d.i32 = int32(value)
	case Uint32T:
		value, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, err
		}

		if value > math.MaxUint32 {
			return nil, fmt.Errorf("Unknown error: the number is not an uint32")
		}

		d.u32 = uint32(value)
	case Int64T:
		value, err := binary.ReadVarint(r)
		if err != nil {
			return nil, err
		}

		d.i64 = int64(value)
	case Uint64T:
		value, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, err
		}

		d.u64 = uint64(value)
	}

	d.number = number
	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

	switch d.GetType() {
	case Int32T:
		return d.i32, nil
	case Uint32T:
		return d.u32, nil
	case Int64T:
		return d.i64, nil
	case Uint64T:
		return d.u64, nil
	}

	return nil, fmt.Errorf("Unkdown error: invalid number type")
}

// ToJSON prepares the data for MarshalJSON
func (d *Number) ToJSON() (interface{}, error) {
	switch d.GetType() {
	case Int32T:
		return d.i32, nil
	case Uint32T:
		return d.u32, nil
	case Int64T:
		return d.i64, nil
	case Uint64T:
		return d.u64, nil
	}

	return nil, fmt.Errorf("Number: unexpected type %d", d.GetType())
}

// Resolve resolves the value
func (d *Number) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	switch d.GetType() {
	case Int32T:
		return d.i32, nil, nil
	case Uint32T:
		return d.u32, nil, nil
	case Int64T:
		return d.i64, nil, nil
	case Uint64T:
		return d.u64, nil, nil
	}

	return nil, nil, fmt.Errorf("Number: unknown error")
}

// ==================================================
// String
// ==================================================

// String is a data handler for the string
type String struct {
	*DataBase

	value  string
	filter *map[string]struct{}
}

var _ Data = (*String)(nil)

// NewString creates a string data handler
func NewString(key string, isRequired bool) *String {
	return &String{
		DataBase: NewDataBase(key, isRequired),
	}
}

// NewStringWithFilter creates a string data handler with filter
func NewStringWithFilter(key string, isRequired bool, filter []string) *String {
	filterPtr := &map[string]struct{}{}
	for _, value := range filter {
		(*filterPtr)[value] = struct{}{}
	}

	return &String{
		DataBase: NewDataBase(key, isRequired),
		filter:   filterPtr,
	}
}

// Prototype creates a prototype String
func (d *String) Prototype() Data {
	return &String{
		DataBase: d.DataBase.Prototype(),
		filter:   d.filter,
	}
}

// Get returns the string value
func (d *String) Get() string {
	return d.value
}

// Set the value of String
func (d *String) Set(data interface{}) error {
	if value, ok := data.(string); ok {
		if d.filter != nil {
			if _, ok := (*d.filter)[value]; !ok {
				return fmt.Errorf("String: %q is not a valid value", value)
			}
		}

		d.value = value
		return d.DataBase.Set(data)
	}

	return fmt.Errorf("String: 'string' is expected but '%T' is found", data)
}

// Encode String
func (d *String) Encode() (interface{}, error) {
	return d.value, nil
}

// Decode String
func (d *String) Decode(data interface{}) (interface{}, error) {
	if err := d.Set(data); err != nil {
		return nil, err
	}

	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

	return d.value, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *String) ToJSON() (interface{}, error) {
	return d.value, nil
}

// Resolve resolves the value
func (d *String) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.value, nil, nil
}

// ==================================================
// Context
// ==================================================

const (
	// ContextKey is the key of context of ISCN object
	ContextKey = "context"
)

// Context is a data handler for the context of ISCN object
type Context struct {
	*Number

	schema  string
	version uint64
}

var _ Data = (*Context)(nil)

// NewContext creates a context of ISCN object with schema name
func NewContext(schema string) *Context {
	return &Context{
		Number: NewNumber(ContextKey, true, Uint64T),
		schema: schema,
	}
}

// Prototype creates a prototype Context
func (d *Context) Prototype() Data {
	return &Context{
		Number: NewNumber(d.GetKey(), d.IsRequired(), d.Number.typ),
		schema: d.schema,
	}
}

// Set the value of Context
func (d *Context) Set(data interface{}) error {
	err := d.Number.Set(data)
	if err != nil {
		return fmt.Errorf("Context: 'uint64' is expected but '%T' is found", data)
	}

	version, err := d.GetUint64()
	if err != nil {
		return fmt.Errorf("Context: %s", err)
	}

	d.version = version
	return d.DataBase.Set(data)
}

// Encode Context
func (d *Context) Encode() (interface{}, error) {
	return d.version, nil
}

// Decode Context
func (d *Context) Decode(data interface{}) (interface{}, error) {
	if err := d.Set(data); err != nil {
		return nil, err
	}

	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

	return d.version, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *Context) ToJSON() (interface{}, error) {
	return d.getSchemaURL(), nil
}

// Resolve resolves the value
func (d *Context) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.getSchemaURL(), nil, nil
}

func (d *Context) getSchemaURL() string {
	// TODO use the real schema path
	return fmt.Sprintf("schema/%s-v%d", d.schema, d.version)
}

// ==================================================
// Cid
// ==================================================

// Cid is a data handler for IPFS CID
type Cid struct {
	*DataBase

	codec uint64
	c     []byte
}

var _ Data = (*Cid)(nil)

// NewCid creates a IPFS CID data handler
func NewCid(key string, isRequired bool, codec uint64) *Cid {
	return &Cid{
		DataBase: NewDataBase(key, isRequired),
		codec:    codec,
	}
}

// Prototype creates a prototype Cid
func (d *Cid) Prototype() Data {
	return &Cid{
		DataBase: d.DataBase.Prototype(),
		codec:    d.codec,
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
func (d *Cid) Set(data interface{}) error {
	if c, ok := data.(cid.Cid); ok {
		if d.codec != 0 && c.Type() != d.codec {
			return fmt.Errorf(
				"Cid: Codec '0x%x' is expected but '0x%x' is found",
				d.codec,
				c.Type())
		}

		d.c = c.Bytes()
		return d.DataBase.Set(data)
	}

	return fmt.Errorf("Cid: 'cid.Cid' is expected but '%T' is found", data)
}

// Encode Cid
func (d *Cid) Encode() (interface{}, error) {
	return d.c, nil
}

// Decode Cid
func (d *Cid) Decode(data interface{}) (interface{}, error) {
	c, ok := data.([]byte)
	if !ok {
		return nil,
			fmt.Errorf("Unknown error during decoding Cid: "+
				"'[]byte' is expected but '%T' is found",
				data,
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
	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

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

// ==================================================
// Timestamp
// ==================================================

const (
	// TimestampPattern is the regexp for specific ISO 8601 datetime string
	TimestampPattern = `^[0-9]{4}` + `-` +
		`(?:1[0-2]|0[1-9])` + `-` +
		`(?:3[01]|0[1-9]|[12][0-9])` + `T` +
		`(?:2[0-3]|[01][0-9])` + `:` +
		`(?:[0-5][0-9])` + `:` +
		`(?:[0-5][0-9])` +
		`(?:Z|[+-](?:2[0-3]|[01][0-9]):(?:[0-5][0-9]))$`
)

// Timestamp is a data handler for a ISO 8601 timestamp string
type Timestamp struct {
	*DataBase

	ts string
}

var _ Data = (*Timestamp)(nil)

// NewTimestamp creates a ISO 8601 timestamp string handler
func NewTimestamp(key string, isRequired bool) *Timestamp {
	return &Timestamp{
		DataBase: NewDataBase(key, isRequired),
	}
}

// Prototype creates a prototype Timestamp
func (d *Timestamp) Prototype() Data {
	return &Timestamp{
		DataBase: d.DataBase.Prototype(),
	}
}

// Set the value of timestamp string
func (d *Timestamp) Set(data interface{}) error {
	if ts, ok := data.(string); ok {
		matched, err := regexp.MatchString(TimestampPattern, ts)
		if err != nil {
			return err
		}

		if !matched {
			return fmt.Errorf("Timestamp: string must in pattern " +
				"YYYY-MM-DDTHH:MM:SS(Z|±HH:MM)")
		}

		d.ts = ts
		return d.DataBase.Set(data)
	}

	return fmt.Errorf("Timestamp: 'string' is expected but '%T' is found", data)
}

// Encode Timestamp
func (d *Timestamp) Encode() (interface{}, error) {
	return d.ts, nil
}

// Decode Timestamp
func (d *Timestamp) Decode(data interface{}) (interface{}, error) {
	ts, ok := data.(string)
	if !ok {
		return nil,
			fmt.Errorf("Unknown error during decoding Timestamp: "+
				"'string' is expected but '%T' is found",
				data,
			)
	}

	matched, err := regexp.MatchString(TimestampPattern, ts)
	if err != nil {
		return nil, err
	}

	if !matched {
		return nil,
			fmt.Errorf("Timestamp: string must in pattern " +
				"YYYY-MM-DDTHH:MM:SS(Z|±HH:MM)",
			)
	}

	d.ts = ts
	if _, err := d.DataBase.Decode(data); err != nil {
		return nil, err
	}

	return ts, nil
}

// ToJSON prepares the data for MarshalJSON
func (d *Timestamp) ToJSON() (interface{}, error) {
	return d.ts, nil
}

// Resolve resolves the value
func (d *Timestamp) Resolve(path []string) (interface{}, []string, error) {
	if len(path) != 0 {
		return nil, nil, fmt.Errorf("Unexpected path elements past %s", path[0])
	}

	return d.ts, nil, nil
}
