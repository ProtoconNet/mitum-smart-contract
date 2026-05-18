//go:build ignore
// +build ignore

// Runtime-only Gno contract example. This file is not part of the normal Go build.
package contract

import "mitum/chain"

type contractError string

func (e contractError) Error() string {
	return string(e)
}

type Meta struct {
	Active bool
	Limit  int64
	Tags   []string
	Flags  map[string]bool
}

type Record struct {
	Name  string
	Value string
	Meta  Meta
}

var initialized bool
var owner string
var value string
var revision int64

var record Record
var flags map[string]bool
var users map[string]Record
var watchers []Record

func Initialize(ctx chain.ContractContext) error {
	if initialized {
		return nil
	}

	owner = ctx.GetSender()
	value = ""
	revision = 0
	record = buildRecord("empty", "", 0, false)
	flags = map[string]bool{
		"initialized": true,
		"created":     false,
		"updated":     false,
	}
	users = map[string]Record{}
	watchers = []Record{}
	initialized = true

	return nil
}

func CreateData(ctx chain.ContractContext, data string) error {
	if !initialized {
		return contractError("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return contractError("only owner can create data")
	}
	if value != "" {
		return contractError("data already exists")
	}

	revision = 1
	applyDataUpdate(ctx, data, true)

	return nil
}

func UpdateData(ctx chain.ContractContext, data string) error {
	if !initialized {
		return contractError("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return contractError("only owner can update data")
	}
	if value == "" {
		return contractError("data does not exist")
	}

	revision = revision + 1
	applyDataUpdate(ctx, data, false)

	return nil
}

func applyDataUpdate(ctx chain.ContractContext, data string, created bool) {
	value = data
	record = buildRecord("record", data, revision, true)

	if flags == nil {
		flags = map[string]bool{}
	}
	flags["initialized"] = initialized
	flags["created"] = created
	flags["updated"] = !created
	flags["owner_exists"] = chain.AccountExists(owner)

	if users == nil {
		users = map[string]Record{}
	}
	users["owner"] = buildRecord(owner, data, revision, true)
	users["latest"] = buildRecord("latest", data, revision+10, !ctx.IsReadOnly())

	watchers = []Record{
		buildRecord("watcher-a", data, revision, true),
		buildRecord("watcher-b", data, revision+100, false),
	}
}

func buildRecord(name string, data string, limit int64, active bool) Record {
	return Record{
		Name:  name,
		Value: data,
		Meta: Meta{
			Active: active,
			Limit:  limit,
			Tags: []string{
				name,
				data,
			},
			Flags: map[string]bool{
				"active": active,
				"empty":  data == "",
			},
		},
	}
}

func IsInitialized(ctx chain.ContractContext) bool {
	return initialized
}

func GetOwner(ctx chain.ContractContext) string {
	return owner
}

func GetValue(ctx chain.ContractContext) string {
	return value
}

func GetRevision(ctx chain.ContractContext) int64 {
	return revision
}

func GetRecord(ctx chain.ContractContext) Record {
	return record
}

func GetRecordName(ctx chain.ContractContext) string {
	return record.Name
}

func GetRecordLimit(ctx chain.ContractContext) int64 {
	return record.Meta.Limit
}

func GetRecordTagAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(record.Meta.Tags) {
		return "", false
	}

	return record.Meta.Tags[index], true
}

func GetRecordFlag(ctx chain.ContractContext, name string) (bool, bool) {
	v, found := record.Meta.Flags[name]
	return v, found
}

func GetFlags(ctx chain.ContractContext) map[string]bool {
	return flags
}

func GetFlag(ctx chain.ContractContext, name string) (bool, bool) {
	v, found := flags[name]
	return v, found
}

func GetUsers(ctx chain.ContractContext) map[string]Record {
	return users
}

func GetUser(ctx chain.ContractContext, name string) (Record, bool) {
	user, found := users[name]
	return user, found
}

func GetUserLimit(ctx chain.ContractContext, name string) (int64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}

	return user.Meta.Limit, true
}

func GetUserTagAt(ctx chain.ContractContext, name string, index int) (string, bool) {
	user, found := users[name]
	if !found {
		return "", false
	}
	if index < 0 || index >= len(user.Meta.Tags) {
		return "", false
	}

	return user.Meta.Tags[index], true
}

func GetWatchers(ctx chain.ContractContext) []Record {
	return watchers
}

func GetWatcherAt(ctx chain.ContractContext, index int) (Record, bool) {
	if index < 0 || index >= len(watchers) {
		return Record{}, false
	}

	return watchers[index], true
}

func GetWatcherLimitAt(ctx chain.ContractContext, index int) (int64, bool) {
	if index < 0 || index >= len(watchers) {
		return 0, false
	}

	return watchers[index].Meta.Limit, true
}

func GetCurrentContract(ctx chain.ContractContext) string {
	return ctx.GetContract()
}

func GetCurrentHeight(ctx chain.ContractContext) int64 {
	return ctx.GetHeight()
}

func DoesAccountExist(ctx chain.ContractContext, addr string) bool {
	return chain.AccountExists(addr)
}

func IsNamedContractAccount(ctx chain.ContractContext, addr string) bool {
	return chain.IsContractAccount(addr)
}

func GetValueIfPresent(ctx chain.ContractContext) (string, bool) {
	if value == "" {
		return "", false
	}

	return value, true
}
