package u256

import (
	"errors"
	onbloc "gno.land/p/onbloc/uint256"
)

type Uint = onbloc.Uint

func NewUint(v uint64) *Uint { return onbloc.NewUint(v) }
func Zero() *Uint { return onbloc.Zero() }
func One() *Uint { return onbloc.One() }
func FromDecimal(v string) (*Uint, error) { return onbloc.FromDecimal(v) }
func MustFromDecimal(v string) *Uint { return onbloc.MustFromDecimal(v) }
func FromHex(v string) (*Uint, error) { return onbloc.FromHex(v) }
func MustFromHex(v string) *Uint { return onbloc.MustFromHex(v) }

func _nativeMulDiv(a string, b string, denominator string) (string, string)
func _nativeSqrt(x string) string

func MulDiv(a, b, denominator *Uint) (*Uint, error) {
	result, code := _nativeMulDiv(a.String(), b.String(), denominator.String())
	if code != "" {
		return nil, errors.New(code)
	}
	return onbloc.MustFromDecimal(result), nil
}

func Sqrt(x *Uint) *Uint {
	return onbloc.MustFromDecimal(_nativeSqrt(x.String()))
}
