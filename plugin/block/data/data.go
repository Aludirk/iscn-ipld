package data

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
// Base
// ==================================================

// Base is the base struct for handling data property
type Base struct {
	isRequired bool
	isDefinded bool

	key string
}

// NewBase creates a base struct for handling data property
func NewBase(key string, isRequired bool) *Base {
	return &Base{
		isRequired: isRequired,
		isDefinded: false,
		key:        key,
	}
}

// Prototype creates a prototype Base
func (b *Base) Prototype() *Base {
	return &Base{
		isRequired: b.isRequired,
		key:        b.key,
	}
}

// IsRequired checks whether the data handler is required
func (b *Base) IsRequired() bool {
	return b.isRequired
}

// IsDefined checks whether the data is well defined
func (b *Base) IsDefined() bool {
	return b.isDefinded
}

// GetKey returns the key of the data property
func (b *Base) GetKey() string {
	return b.key
}

// MarkDefined mark the data is defined.
func (b *Base) MarkDefined() {
	b.isDefinded = true
}
