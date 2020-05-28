package data

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

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
	*Base

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
		Base: NewBase(key, isRequired),
		typ:  typ,
	}
}

// Prototype creates a prototype Number
func (d *Number) Prototype() Data {
	return &Number{
		Base: d.Base.Prototype(),
		typ:  d.typ,
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
func (d *Number) Set(obj interface{}) error {
	switch d.GetType() {
	case Int32T:
		var value int32
		switch v := obj.(type) {
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
			return fmt.Errorf("Number: 'int32' is expected but '%T' is found", obj)
		}

		buffer := make([]byte, binary.MaxVarintLen32)
		n := binary.PutVarint(buffer, int64(value))
		d.number = buffer[:n]
		d.i32 = value
	case Uint32T:
		var value uint32
		switch v := obj.(type) {
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
			return fmt.Errorf("Number: 'uint32' is expected but '%T' is found", obj)
		}

		buffer := make([]byte, binary.MaxVarintLen32)
		n := binary.PutUvarint(buffer, uint64(value))
		d.number = buffer[:n]
		d.u32 = value
	case Int64T:
		var value int64
		switch v := obj.(type) {
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
			return fmt.Errorf("Number: 'int64' is expected but '%T' is found", obj)
		}

		buffer := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buffer, value)
		d.number = buffer[:n]
		d.i64 = value
	case Uint64T:
		var value uint64
		switch v := obj.(type) {
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
			return fmt.Errorf("Number: 'uint64' is expected but '%T' is found", obj)
		}

		buffer := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buffer, value)
		d.number = buffer[:n]
		d.u64 = value
	default:
		return fmt.Errorf("Number: unknown error")
	}

	d.Base.MarkDefined()
	return nil
}

// Encode Number
func (d *Number) Encode() (interface{}, error) {
	return d.number, nil
}

// Decode Number
func (d *Number) Decode(obj interface{}) (interface{}, error) {
	number, ok := obj.([]byte)
	if !ok {
		return nil,
			fmt.Errorf("Unknown error during decoding number: "+
				"'[]byte' is expected but '%T' is found",
				obj,
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
	d.Base.MarkDefined()

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

	return nil, fmt.Errorf("Unknown error: invalid number type")
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
