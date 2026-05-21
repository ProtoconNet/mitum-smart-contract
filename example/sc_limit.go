//go:build ignore
// +build ignore

// Runtime-only Gno contract for manual payload limit checks.
package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.WriteContext, seed string) error {
	value = seed
	return nil
}

func Store(ctx chain.WriteContext, next string) error {
	value = next
	return nil
}

func GetValue(ctx chain.QueryContext) string {
	return value
}
