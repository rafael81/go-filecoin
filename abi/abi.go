package abi

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/filecoin-project/go-filecoin/types"
)

// ErrInvalidType is returned when processing a zero valued 'Type' (aka Invalid)
var ErrInvalidType = fmt.Errorf("invalid type")

// Type represents a type that can be passed through the filecoin ABI
type Type uint64

const (
	// Invalid is the default value for 'Type' and represents an errorneously set type.
	Invalid = Type(iota)
	// Address is a types.Address
	Address
	// Integer is a *big.Int
	Integer
	// Bytes is a []byte
	Bytes
	// String is a string
	String
)

func (t Type) String() string {
	switch t {
	case Invalid:
		return "<invalid>"
	case Address:
		return "types.Address"
	case Integer:
		return "*big.Int"
	case Bytes:
		return "[]byte"
	case String:
		return "string"
	default:
		return "<unknown type>"
	}
}

// Value pairs a go value with its ABI type
type Value struct {
	Type Type
	Val  interface{}
}

type typeError struct {
	exp interface{}
	got interface{}
}

func (ate typeError) Error() string {
	return fmt.Sprintf("expected type %T, got %T", ate.exp, ate.got)
}

// Serialize serializes the value into raw bytes. Only works on valid supported types.
func (av *Value) Serialize() ([]byte, error) {
	switch av.Type {
	case Invalid:
		return nil, ErrInvalidType
	case Address:
		addr, ok := av.Val.(types.Address)
		if !ok {
			return nil, &typeError{types.Address(""), av.Val}
		}
		return []byte(addr), nil
	case Integer:
		intgr, ok := av.Val.(*big.Int)
		if !ok {
			return nil, &typeError{&big.Int{}, av.Val}
		}
		return intgr.Bytes(), nil
	case Bytes:
		b, ok := av.Val.([]byte)
		if !ok {
			return nil, &typeError{[]byte{}, av.Val}
		}
		return b, nil
	case String:
		s, ok := av.Val.(string)
		if !ok {
			return nil, &typeError{"", av.Val}
		}

		return []byte(s), nil
	default:
		return nil, fmt.Errorf("unrecognized Type: %d", av.Type)
	}
}

// ToValues converts from a slice of go abi-compatible values to abi values.
// empty slices are normalized to nil
func ToValues(i []interface{}) ([]*Value, error) {
	if len(i) == 0 {
		return nil, nil
	}

	out := make([]*Value, 0, len(i))
	for _, v := range i {
		switch v := v.(type) {
		case types.Address:
			out = append(out, &Value{Type: Address, Val: v})
		case *big.Int:
			out = append(out, &Value{Type: Integer, Val: v})
		case []byte:
			out = append(out, &Value{Type: Bytes, Val: v})
		case string:
			out = append(out, &Value{Type: String, Val: v})
		default:
			return nil, fmt.Errorf("unsupported type: %T", v)
		}
	}
	return out, nil
}

// FromValues converts from a slice of abi values to the go type representation
// of them. empty slices are normalized to nil
func FromValues(vals []*Value) []interface{} {
	if len(vals) == 0 {
		return nil
	}

	out := make([]interface{}, 0, len(vals))
	for _, v := range vals {
		out = append(out, v.Val)
	}
	return out
}

// Deserialize converts the given bytes to the requested type and returns an
// ABI Value for it.
func Deserialize(data []byte, t Type) (*Value, error) {
	switch t {
	case Invalid:
		return nil, ErrInvalidType
	case Address:
		return &Value{
			Type: t,
			Val:  types.Address(data),
		}, nil
	case Integer:
		return &Value{
			Type: t,
			Val:  big.NewInt(0).SetBytes(data),
		}, nil
	case Bytes:
		return &Value{
			Type: t,
			Val:  data,
		}, nil
	case String:
		return &Value{
			Type: t,
			Val:  string(data),
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized Type: %d", t)
	}
}

var typeTable = map[Type]reflect.Type{
	Address: reflect.TypeOf(types.Address("")),
	Integer: reflect.TypeOf(&big.Int{}),
	Bytes:   reflect.TypeOf([]byte{}),
	String:  reflect.TypeOf(string("")),
}

// TypeMatches returns whether or not 'val' is the go type expected for the given ABI type
func TypeMatches(t Type, val reflect.Type) bool {
	rt, ok := typeTable[t]
	if !ok {
		return false
	}
	return rt == val
}
