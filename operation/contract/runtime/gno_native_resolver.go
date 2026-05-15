package runtime

import (
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/pkg/errors"
)

func CombineNativeResolvers(resolvers ...gno.NativeResolver) gno.NativeResolver {
	return func(pkgPath string, name gno.Name) func(m *gno.Machine) {
		for _, resolver := range resolvers {
			if resolver == nil {
				continue
			}

			if fn := resolver(pkgPath, name); fn != nil {
				return fn
			}
		}

		return nil
	}
}

func MitumNativeResolver(pkgPath string, name gno.Name) func(m *gno.Machine) {
	if pkgPath != MitumChainPackagePath {
		return nil
	}

	switch string(name) {
	case "AccountExists":
		return nativeAccountExists
	case "IsContractAccount":
		return nativeIsContractAccount
	default:
		return nil
	}
}

func mustExecutionContext(m *gno.Machine) *ExecutionContext {
	ctx, ok := m.Context.(*ExecutionContext)
	if !ok || ctx == nil {
		panic("mitum execution context is missing from machine")
	}

	if err := ctx.Validate(); err != nil {
		panic(err)
	}

	return ctx
}

func machineStringArg(m *gno.Machine, argIndex int) string {
	b := m.LastBlock()

	var out string
	rv := reflect.ValueOf(&out).Elem()

	tv := b.GetPointerTo(nil, gno.NewValuePathBlock(1, uint16(argIndex), "")).TV
	tv.DeepFill(m.Store)
	gno.Gno2GoValue(tv, rv)

	return out
}

func pushBoolResult(m *gno.Machine, v bool) {
	m.PushValue(
		gno.Go2GnoValue(
			m.Alloc,
			m.Store,
			reflect.ValueOf(&v).Elem(),
		),
	)
}

func nativeAccountExists(m *gno.Machine) {
	ctx := mustExecutionContext(m)
	addr := machineStringArg(m, 0)

	ok, err := ctx.AccountReader.AccountExists(addr)
	if err != nil {
		panic(errors.Wrap(err, "AccountExists native call failed"))
	}

	pushBoolResult(m, ok)
}

func nativeIsContractAccount(m *gno.Machine) {
	ctx := mustExecutionContext(m)
	addr := machineStringArg(m, 0)

	ok, err := ctx.ContractReader.IsContractAccount(addr)
	if err != nil {
		panic(errors.Wrap(err, "IsContractAccount native call failed"))
	}

	pushBoolResult(m, ok)
}
