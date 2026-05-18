package runtime

import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"

const mitumNativeFlatGasBase int64 = 3000

func init() {
	// Provisional flat costs for app-specific Mitum host ABI natives.
	gno.RegisterNativeGas(
		MitumChainPackagePath,
		gno.Name("AccountExists"),
		&gno.NativeGasInfo{
			Base:      mitumNativeFlatGasBase,
			SlopeIdx:  -1,
			SlopeKind: gno.SizeFlat,
		},
	)

	gno.RegisterNativeGas(
		MitumChainPackagePath,
		gno.Name("IsContractAccount"),
		&gno.NativeGasInfo{
			Base:      mitumNativeFlatGasBase,
			SlopeIdx:  -1,
			SlopeKind: gno.SizeFlat,
		},
	)
}
