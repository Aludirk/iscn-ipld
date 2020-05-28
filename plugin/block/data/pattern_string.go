package data

import (
	"fmt"
	"net/url"
	"regexp"
)

// ==================================================
// PatternString
// ==================================================

// PatternString is a data handler for a string that should match a given pattern
type PatternString struct {
	*Base

	value   *String
	pattern *regexp.Regexp
}

var _ Data = (*PatternString)(nil)

// NewPatternString creates a pattern string handler
func NewPatternString(key string, isRequired bool, expr string) *PatternString {
	pattern, err := regexp.Compile(expr)
	if err != nil {
		panic(fmt.Sprintf("PatternString: invalid expression %s (%s)", expr, err))
	}

	return &PatternString{
		Base:    NewBase(key, isRequired),
		value:   NewString("", false),
		pattern: pattern,
	}
}

// Prototype creates a prototype PatternString
func (d *PatternString) Prototype() Data {
	return &PatternString{
		Base:    d.Base.Prototype(),
		value:   NewString(d.value.GetKey(), d.value.IsRequired()),
		pattern: d.pattern,
	}
}

// Get returns the string value
func (d *PatternString) Get() string {
	return d.value.Get()
}

// Set the value of PatternString string
func (d *PatternString) Set(obj interface{}) error {
	if err := d.value.Set(obj); err != nil {
		return err
	}

	if !d.pattern.MatchString(d.Get()) {
		return fmt.Errorf(
			"PatternString: string must match the pattern %s",
			d.pattern.String(),
		)
	}

	d.Base.MarkDefined()
	return nil
}

// Encode PatternString
func (d *PatternString) Encode() (interface{}, error) {
	return d.value.Encode()
}

// Decode PatternString
func (d *PatternString) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.Get(), nil
}

// ToJSON prepares the data for MarshalJSON
func (d *PatternString) ToJSON() (interface{}, error) {
	return d.value.ToJSON()
}

// Resolve resolves the value
func (d *PatternString) Resolve(path []string) (interface{}, []string, error) {
	return d.value.Resolve(path)
}

// ==================================================
// Timestamp
// ==================================================

// Timestamp is a data handler for a ISO 8601 timestamp string
type Timestamp struct {
	*PatternString
}

var _ Data = (*Timestamp)(nil)

// NewTimestamp creates a ISO 8601 timestamp string handler
func NewTimestamp(key string, isRequired bool) *Timestamp {
	base := NewPatternString(
		key,
		isRequired,
		`^[0-9]{4}`+`-`+
			`(?:1[0-2]|0[1-9])`+`-`+
			`(?:3[01]|0[1-9]|[12][0-9])`+`T`+
			`(?:2[0-3]|[01][0-9])`+`:`+
			`(?:[0-5][0-9])`+`:`+
			`(?:[0-5][0-9])`+
			`(?:Z|[+-](?:2[0-3]|[01][0-9]):(?:[0-5][0-9]))$`,
	)

	return &Timestamp{
		PatternString: base,
	}
}

// Prototype creates a prototype Timestamp
func (d *Timestamp) Prototype() Data {
	return &Timestamp{
		PatternString: NewPatternString(d.GetKey(), d.IsRequired(), d.pattern.String()),
	}
}

// ==================================================
// LikeCoinChainID
// ==================================================

// LikeCoinChainID is a data handler for the LikeCoin chain ID
type LikeCoinChainID struct {
	*PatternString
}

var _ Data = (*LikeCoinChainID)(nil)

// NewLikeCoinChainID creates a LikeCoin chain ID handler
func NewLikeCoinChainID(key string, isRequired bool) *LikeCoinChainID {
	base := NewPatternString(
		key,
		isRequired,
		`lcc://id/cosmos1[02-9ac-hj-np-z]{38}$`,
	)

	return &LikeCoinChainID{
		PatternString: base,
	}
}

// Prototype creates a prototype LikeCoinChainID
func (d *LikeCoinChainID) Prototype() Data {
	return &LikeCoinChainID{
		PatternString: NewPatternString(d.GetKey(), d.IsRequired(), d.pattern.String()),
	}
}

// ==================================================
// Hash
// ==================================================

// Hash is a data handler for the URL of hash
type Hash struct {
	*PatternString
}

var _ Data = (*Hash)(nil)

// NewHash creates a hash URL handler
func NewHash(key string, isRequired bool) *Hash {
	base := NewPatternString(
		key,
		isRequired,
		`^hash://[^/]+/[0-9a-f]+$`,
	)

	return &Hash{
		PatternString: base,
	}
}

// Prototype creates a prototype Hash
func (d *Hash) Prototype() Data {
	return &Hash{
		PatternString: NewPatternString(d.GetKey(), d.IsRequired(), d.pattern.String()),
	}
}

// ==================================================
// URL
// ==================================================

// URL is a data handler for an URL string
type URL struct {
	*Base

	value *String
}

var _ Data = (*URL)(nil)

// NewURL creates an URL string handler
func NewURL(key string, isRequired bool) *URL {
	return &URL{
		Base:  NewBase(key, isRequired),
		value: NewString("", false),
	}
}

// Prototype creates a prototype URL
func (d *URL) Prototype() Data {
	return &URL{
		Base:  d.Base.Prototype(),
		value: NewString(d.value.GetKey(), d.value.IsRequired()),
	}
}

// Get returns the string value
func (d *URL) Get() string {
	return d.value.Get()
}

// Set the value of URL
func (d *URL) Set(obj interface{}) error {
	if err := d.value.Set(obj); err != nil {
		return err
	}

	value, err := url.Parse(d.Get())
	if err != nil {
		return fmt.Errorf("URL: invalid URL %s", err)
	}

	matched, err := regexp.MatchString(`^(?:https?|ftp)$`, value.Scheme)
	if err != nil {
		return fmt.Errorf("URL: cannot check the scheme %q", value.Scheme)
	}

	if !matched {
		return fmt.Errorf("URL: scheme must match the pattern \"(https?|ftp)\"")
	}

	d.Base.MarkDefined()
	return nil
}

// Encode URL
func (d *URL) Encode() (interface{}, error) {
	return d.value.Encode()
}

// Decode URL
func (d *URL) Decode(obj interface{}) (interface{}, error) {
	if err := d.Set(obj); err != nil {
		return nil, err
	}

	d.Base.MarkDefined()
	return d.Get(), nil
}

// ToJSON prepares the data for MarshalJSON
func (d *URL) ToJSON() (interface{}, error) {
	return d.value.ToJSON()
}

// Resolve resolves the value
func (d *URL) Resolve(path []string) (interface{}, []string, error) {
	return d.value.Resolve(path)
}
