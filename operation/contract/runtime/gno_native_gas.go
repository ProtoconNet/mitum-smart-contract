package runtime

import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

const (
	// These are defensive provisional tiers, not benchmark-calibrated values.
	// AccountExists and IsContractAccount each decode an address and read one
	// state key. BalanceOf can validate two args and read account, currency
	// design, and balance keys.
	mitumNativeSingleLookupGasBase int64 = 3000
	mitumNativeTripleLookupGasBase int64 = mitumNativeSingleLookupGasBase * 3
)

func init() {
	// Provisional flat costs for app-specific Mitum host ABI natives.
	registerMitumNativeFlatGas("AccountExists", mitumNativeSingleLookupGasBase)
	registerMitumNativeFlatGas("IsContractAccount", mitumNativeSingleLookupGasBase)
	registerMitumNativeFlatGas("BalanceOf", mitumNativeTripleLookupGasBase)
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
