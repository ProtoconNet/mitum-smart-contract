package runtime

import (
	"math/big"
	"math/rand"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
)

func TestUint256NativeMulDivBoundaries(t *testing.T) {
	for _, tc := range []struct {
		a, b, d, result, code string
	}{
		{"6", "7", "4", "10", ""},
		{u256Max, u256Max, u256Max, u256Max, ""},
		{"1", "1", "0", "", Uint256MulDivDenominatorZero},
		{u256Max, u256Max, "1", "", Uint256MulDivResultOverflow},
		{"01", "1", "1", "", Uint256InvalidNativeInput},
	} {
		result, code := uint256MulDivDecimal(tc.a, tc.b, tc.d)
		if result != tc.result || code != tc.code {
			t.Fatalf("MulDiv(%s,%s,%s) = (%q,%q), want (%q,%q)", tc.a, tc.b, tc.d, result, code, tc.result, tc.code)
		}
	}
}

func TestUint256NativeDeterministicBigIntDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(4256))
	for i := 0; i < 64; i++ {
		a := new(big.Int).Rand(rng, maxUint256Big)
		b := new(big.Int).Rand(rng, maxUint256Big)
		d := new(big.Int).Rand(rng, maxUint256Big)
		if d.Sign() == 0 {
			d.SetInt64(1)
		}
		want := new(big.Int).Div(new(big.Int).Mul(a, b), d)
		result, code := uint256MulDivDecimal(a.String(), b.String(), d.String())
		if want.Cmp(maxUint256Big) > 0 {
			if code != Uint256MulDivResultOverflow {
				t.Fatalf("vector %d expected overflow, got (%q,%q)", i, result, code)
			}
			continue
		}
		if code != "" || result != want.String() {
			t.Fatalf("vector %d got (%q,%q), want %s", i, result, code, want)
		}

		x := new(big.Int).Rand(rng, maxUint256Big)
		sqrt, ok := uint256SqrtDecimal(x.String())
		if !ok || sqrt != new(big.Int).Sqrt(x).String() {
			t.Fatalf("sqrt vector %d got (%q,%v)", i, sqrt, ok)
		}
	}
}

func TestUint256NativeRejectsNonCanonicalInputs(t *testing.T) {
	for _, input := range []string{"", "00", "01", "-1", "+1", " 1", "1 ", "0x1", u256Modulus} {
		if _, ok := parseCanonicalUint256(input); ok {
			t.Fatalf("accepted non-canonical uint256 %q", input)
		}
	}
}

func TestUint256NativeIdentityIsExact(t *testing.T) {
	for _, name := range []gno.Name{"_nativeMulDiv", "_nativeSqrt"} {
		if Uint256NativeResolver(Uint256PackagePath, name) == nil {
			t.Fatalf("missing native resolver for %s.%s", Uint256PackagePath, name)
		}
		for _, path := range []string{"mitum/math/u256/v1", "mitum/math/v1", "mitum/math/v1/u256/subpackage", "mitum/math/v1/u2562"} {
			if Uint256NativeResolver(path, name) != nil {
				t.Fatalf("native resolver accepted %s.%s", path, name)
			}
		}
	}
	for _, name := range []string{"v1", "uint256", "other"} {
		pkg := &gnostd.MemPackage{Name: name, Path: Uint256PackagePath, Type: gno.MPStdlibProd, Files: []*gnostd.MemFile{{Name: "u256.gno", Body: "package " + name}}}
		if err := validateGnoPurePackageIdentity(pkg); err == nil {
			t.Fatalf("accepted package name %q at %q", name, Uint256PackagePath)
		}
	}
}
