package runtime

import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

const (
	// These are defensive provisional tiers, not benchmark-calibrated values.
	// AccountExists and IsContractAccount each decode an address and read one
	// state key. BalanceOf can validate two args and read account, currency
	// design, and balance keys. SHA3Sum256 is pure CPU work, charged as a
	// small dispatch/hash base plus exact per-input-byte cost using Gno's
	// per-KiB native slope convention.
	mitumNativeSingleLookupGasBase  int64 = 3000
	mitumNativeTripleLookupGasBase  int64 = mitumNativeSingleLookupGasBase * 3
	mitumNativeSHA3Sum256GasBase    int64 = 1000
	mitumNativeSHA3Sum256GasPerByte int64 = 2
	// Uint256 costs are provisional flat dispatch tiers. They are deliberately
	// below the fixed query budget so a single bounded 256-bit operation works;
	// repeated calls still consume the shared machine meter deterministically.
	uint256NativeMulDivGasBase int64 = 1
	uint256NativeSqrtGasBase   int64 = 1
)

func init() {
	// Provisional costs for app-specific Mitum host ABI natives.
	registerMitumNativeFlatGas("AccountExists", mitumNativeSingleLookupGasBase)
	registerMitumNativeFlatGas("IsContractAccount", mitumNativeSingleLookupGasBase)
	registerMitumNativeFlatGas("BalanceOf", mitumNativeTripleLookupGasBase)
	registerMitumNativeStringLinearGas("SHA3Sum256", mitumNativeSHA3Sum256GasBase, mitumNativeSHA3Sum256GasPerByte)
	registerUint256NativeFlatGas("_nativeMulDiv", uint256NativeMulDivGasBase)
	registerUint256NativeFlatGas("_nativeSqrt", uint256NativeSqrtGasBase)
}

func registerUint256NativeFlatGas(name string, base int64) {
	gno.RegisterNativeGas(Uint256PackagePath, gno.Name(name), &gno.NativeGasInfo{Base: base, SlopeIdx: -1, SlopeKind: gno.SizeFlat})
}

func registerMitumNativeFlatGas(name string, base int64) {
	gno.RegisterNativeGas(
		MitumChainPackagePath,
		gno.Name(name),
		&gno.NativeGasInfo{
			Base:      base,
			SlopeIdx:  -1,
			SlopeKind: gno.SizeFlat,
		},
	)
}

func registerMitumNativeStringLinearGas(name string, base int64, perByte int64) {
	gno.RegisterNativeGas(
		MitumChainPackagePath,
		gno.Name(name),
		&gno.NativeGasInfo{
			Base:      base,
			Slope:     perByte * 1024,
			SlopeIdx:  0,
			SlopeKind: gno.SizeLenString,
		},
	)
}

func mitumNativeSHA3Sum256Gas(inputBytes int) int64 {
	return mitumNativeSHA3Sum256GasBase + mitumNativeSHA3Sum256GasPerByte*int64(inputBytes)
}
