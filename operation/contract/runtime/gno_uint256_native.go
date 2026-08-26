package runtime

import (
	"math/big"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const (
	Uint256NativeSemanticsVersion = "mitum-u256-native-v2"
	Uint256NativeGasSchedule      = "mitum-u256-provisional-gas-v1"
	Uint256CanonicalDecimalPolicy = "mitum-u256-canonical-decimal-v1"
	Uint256CanonicalDecimalError  = "u256: non-canonical decimal"
	Uint256WrapperSourceSHA256    = "43980a468b2a107b64c3efa4ced7568b5f21a8f012dbcdd22fd94dae03cee899"
	Uint256MulDivDenominatorZero  = "u256: denominator is zero"
	Uint256MulDivResultOverflow   = "u256: result overflow"
	Uint256InvalidNativeInput     = "u256: invalid native input"
)

var maxUint256Big = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

func Uint256NativeResolver(pkgPath string, name gno.Name) func(*gno.Machine) {
	if pkgPath != Uint256PackagePath {
		return nil
	}
	switch string(name) {
	case "_nativeCanonicalToHex":
		return nativeUint256CanonicalToHex
	case "_nativeMulDiv":
		return nativeUint256MulDiv
	case "_nativeSqrt":
		return nativeUint256Sqrt
	default:
		return nil
	}
}

func nativeUint256CanonicalToHex(m *gno.Machine) {
	hex, code := uint256CanonicalToHex(machineStringArg(m, 0))
	pushStringResult(m, hex)
	pushStringResult(m, code)
}

func nativeUint256MulDiv(m *gno.Machine) {
	decimal, code := uint256MulDivDecimal(machineStringArg(m, 0), machineStringArg(m, 1), machineStringArg(m, 2))
	hex := ""
	if code == "" {
		x, _ := parseCanonicalUint256(decimal)
		hex = canonicalUint256Hex(x)
	}
	pushStringResult(m, hex)
	pushStringResult(m, decimal)
	pushStringResult(m, code)
}

func nativeUint256Sqrt(m *gno.Machine) {
	decimal, ok := uint256SqrtDecimal(machineStringArg(m, 0))
	if !ok {
		pushStringResult(m, "")
		pushStringResult(m, "")
		pushStringResult(m, Uint256CanonicalDecimalError)
		return
	}
	x, _ := parseCanonicalUint256(decimal)
	pushStringResult(m, canonicalUint256Hex(x))
	pushStringResult(m, decimal)
	pushStringResult(m, "")
}

func uint256CanonicalToHex(input string) (string, string) {
	x, ok := parseCanonicalUint256(input)
	if !ok {
		return "", Uint256CanonicalDecimalError
	}
	return canonicalUint256Hex(x), ""
}

func canonicalUint256Hex(x *big.Int) string { return "0x" + x.Text(16) }

func uint256MulDivDecimal(a, b, denominator string) (string, string) {
	x, ok := parseCanonicalUint256(a)
	if !ok {
		return "", Uint256CanonicalDecimalError
	}
	y, ok := parseCanonicalUint256(b)
	if !ok {
		return "", Uint256CanonicalDecimalError
	}
	d, ok := parseCanonicalUint256(denominator)
	if !ok {
		return "", Uint256CanonicalDecimalError
	}
	if d.Sign() == 0 {
		return "", Uint256MulDivDenominatorZero
	}
	result := new(big.Int).Div(new(big.Int).Mul(x, y), d)
	if result.Cmp(maxUint256Big) > 0 {
		return "", Uint256MulDivResultOverflow
	}
	return result.String(), ""
}

func uint256SqrtDecimal(input string) (string, bool) {
	x, ok := parseCanonicalUint256(input)
	if !ok {
		return "", false
	}
	return new(big.Int).Sqrt(x).String(), true
}

func parseCanonicalUint256(input string) (*big.Int, bool) {
	if input == "" || len(input) > 78 || (len(input) > 1 && input[0] == '0') {
		return nil, false
	}
	if strings.IndexFunc(input, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return nil, false
	}
	x, ok := new(big.Int).SetString(input, 10)
	if !ok || x.Sign() < 0 || x.Cmp(maxUint256Big) > 0 || x.String() != input {
		return nil, false
	}
	return x, true
}
