//go:build ignore
// +build ignore

// Runtime-only Gno contract example. This file is not part of the normal Go build.
package contract

import (
	"fmt"

	"mitum/chain"
)

var initialized bool
var owner string
var value string
var revision int64

func Initialize(ctx chain.ContractContext) error {
	if initialized {
		return nil
	}

	owner = ctx.GetSender()
	value = ""
	revision = 0
	initialized = true

	return nil
}

func CreateData(ctx chain.ContractContext, data string) error {
	if !initialized {
		return fmt.Errorf("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return fmt.Errorf("only owner can create data")
	}
	if value != "" {
		return fmt.Errorf("data already exists")
	}

	value = data
	revision = 1

	return nil
}

func UpdateData(ctx chain.ContractContext, data string) error {
	if !initialized {
		return fmt.Errorf("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return fmt.Errorf("only owner can update data")
	}
	if value == "" {
		return fmt.Errorf("data does not exist")
	}

	value = data
	revision = revision + 1

	return nil
}

func GetValue(ctx chain.ContractContext) string {
	return value
}

func GetRevision(ctx chain.ContractContext) int64 {
	return revision
}

func GetValueIfPresent(ctx chain.ContractContext) (string, bool) {
	if value == "" {
		return "", false
	}

	return value, true
}
