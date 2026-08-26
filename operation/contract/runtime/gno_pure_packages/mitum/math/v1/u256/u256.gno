package u256

import (
	"errors"
	onbloc "gno.land/p/onbloc/uint256"
)

type Uint = onbloc.Uint

const CanonicalDecimalError = "u256: non-canonical decimal"

func NewUint(v uint64) *Uint { return onbloc.NewUint(v) }
func Zero() *Uint { return onbloc.Zero() }
func One() *Uint { return onbloc.One() }
func FromDecimal(v string) (*Uint, error) { return onbloc.FromDecimal(v) }
func MustFromDecimal(v string) *Uint { return onbloc.MustFromDecimal(v) }
func FromHex(v string) (*Uint, error) { return onbloc.FromHex(v) }
func MustFromHex(v string) *Uint { return onbloc.MustFromHex(v) }

func hasCanonicalDecimalSyntax(v string) bool {
	return len(v) > 0 && len(v) <= 78
}

func IsCanonicalDecimal(v string) bool {
	_, code := _nativeCanonicalToHex(v)
	return code == ""
}

func FromCanonicalDecimal(v string) (*Uint, error) {
	hex, code := _nativeCanonicalToHex(v)
	if code != "" {
		return nil, errors.New(code)
	}
	return onbloc.FromHex(hex)
}

func MustFromCanonicalDecimal(v string) *Uint {
	parsed, err := FromCanonicalDecimal(v)
	if err != nil {
		panic(CanonicalDecimalError)
	}
	return parsed
}

func ToCanonicalDecimal(v *Uint) string { return v.String() }

func _nativeCanonicalToHex(v string) (string, string)
func _nativeMulDiv(a string, b string, denominator string) (string, string, string)
func _nativeSqrt(x string) (string, string, string)

func MulDiv(a, b, denominator *Uint) (*Uint, error) {
	hex, _, code := _nativeMulDiv(a.String(), b.String(), denominator.String())
	if code != "" {
		return nil, errors.New(code)
	}
	return onbloc.MustFromHex(hex), nil
}

func Sqrt(x *Uint) *Uint {
	hex, _, code := _nativeSqrt(x.String())
	if code != "" {
		panic(code)
	}
	return onbloc.MustFromHex(hex)
}

func MulDivCanonicalDecimal(a, b, denominator string) (string, error) {
	_, decimal, code := _nativeMulDiv(a, b, denominator)
	if code != "" {
		return "", errors.New(code)
	}
	return decimal, nil
}

func SqrtCanonicalDecimal(x string) (string, error) {
	_, decimal, code := _nativeSqrt(x)
	if code != "" {
		return "", errors.New(code)
	}
	return decimal, nil
}
