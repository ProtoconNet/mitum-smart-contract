//go:build ignore
// +build ignore

// Runtime-only Gno contract example. This file is not part of the normal Go build.
package contract

import (
	"mitum/chain"
	"mitum/math/v1/u256"
)

// Persistent values stay canonical decimal strings. Uint values are local calculations only.
var total string
var lastSqrt string

func Initialize(ctx chain.WriteContext, initial string) error {
	value, err := u256.FromCanonicalDecimal(initial)
	if err != nil {
		return err
	}

	total = u256.ToCanonicalDecimal(value)
	lastSqrt = "0"
	return nil
}

func Add(ctx chain.WriteContext, amount string) error {
	current, err := u256.FromCanonicalDecimal(total)
	if err != nil {
		return err
	}
	delta, err := u256.FromCanonicalDecimal(amount)
	if err != nil {
		return err
	}

	// Convert back before crossing the state boundary.
	total = u256.ToCanonicalDecimal(new(u256.Uint).Add(current, delta))
	return nil
}

func ApplyRatio(ctx chain.WriteContext, numerator string, denominator string) error {
	// Decimal-native APIs avoid unnecessary Uint conversion at state boundaries.
	next, err := u256.MulDivCanonicalDecimal(total, numerator, denominator)
	if err != nil {
		return err
	}

	total = next
	return nil
}

func StoreSqrt(ctx chain.WriteContext) error {
	next, err := u256.SqrtCanonicalDecimal(total)
	if err != nil {
		return err
	}

	lastSqrt = next
	return nil
}

func Reset(ctx chain.WriteContext, next string) error {
	value, err := u256.FromCanonicalDecimal(next)
	if err != nil {
		return err
	}

	total = u256.ToCanonicalDecimal(value)
	lastSqrt = "0"
	return nil
}

func GetTotal(ctx chain.QueryContext) string {
	return total
}

func GetLastSqrt(ctx chain.QueryContext) string {
	return lastSqrt
}

func PreviewRatio(ctx chain.QueryContext, numerator string, denominator string) (string, bool) {
	result, err := u256.MulDivCanonicalDecimal(total, numerator, denominator)
	if err != nil {
		return "", false
	}
	return result, true
}

func PreviewRatioWithError(ctx chain.QueryContext, numerator string, denominator string) string {
	result, err := u256.MulDivCanonicalDecimal(total, numerator, denominator)
	if err != nil {
		return err.Error()
	}
	return result
}

func PreviewSqrt(ctx chain.QueryContext) string {
	// Query calculations return canonical strings without mutating state.
	result, err := u256.SqrtCanonicalDecimal(total)
	if err != nil {
		return "0"
	}
	return result
}
