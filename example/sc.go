//go:build ignore
// +build ignore

// Runtime-only Gno contract example. This file is not part of the normal Go build.
package contract

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"mitum/chain"
	"strconv"
	"strings"
	"unicode/utf8"
)

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
var blockTime int64

var record Record
var flags map[string]bool
var users map[string]Record
var watchers []Record

func Initialize(ctx chain.WriteContext, initialValue string, initialLimit int64) error {
	if initialized {
		return nil
	}

	owner = ctx.GetSender()
	value = initialValue
	revision = 0
	blockTime = ctx.GetBlockTime()
	record = buildRecord("initial", initialValue, initialLimit, false)
	flags = map[string]bool{
		"initialized": true,
		"created":     false,
		"updated":     false,
	}
	users = map[string]Record{
		"owner": buildRecord(owner, initialValue, initialLimit, true),
	}
	watchers = []Record{
		buildRecord("watcher-seed", initialValue, initialLimit, false),
	}
	initialized = true

	return nil
}

func CreateData(ctx chain.WriteContext, data string) error {
	if !initialized {
		return contractError("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return contractError("only owner can create data")
	}
	if flags["created"] {
		return contractError("data already exists")
	}

	revision = 1
	applyDataUpdate(ctx, data, true)

	return nil
}

func UpdateData(ctx chain.WriteContext, data string) error {
	if !initialized {
		return contractError("contract is not initialized")
	}
	if ctx.GetSender() != owner {
		return contractError("only owner can update data")
	}
	if !flags["created"] {
		return contractError("data does not exist")
	}

	revision = revision + 1
	applyDataUpdate(ctx, data, false)

	return nil
}

func applyDataUpdate(ctx chain.WriteContext, data string, created bool) {
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

func IsInitialized(ctx chain.QueryContext) bool {
	return initialized
}

func GetOwner(ctx chain.QueryContext) string {
	return owner
}

func GetValue(ctx chain.QueryContext) string {
	return value
}

func GetValueSHA3(ctx chain.QueryContext) string {
	return chain.SHA3Sum256(value)
}

func GetRevision(ctx chain.QueryContext) int64 {
	return revision
}

func GetRecord(ctx chain.QueryContext) Record {
	return record
}

func GetRecordName(ctx chain.QueryContext) string {
	return record.Name
}

func GetRecordLimit(ctx chain.QueryContext) int64 {
	return record.Meta.Limit
}

func GetRecordTagAt(ctx chain.QueryContext, index int) (string, bool) {
	if index < 0 || index >= len(record.Meta.Tags) {
		return "", false
	}

	return record.Meta.Tags[index], true
}

func GetRecordFlag(ctx chain.QueryContext, name string) (bool, bool) {
	v, found := record.Meta.Flags[name]
	return v, found
}

func GetFlags(ctx chain.QueryContext) map[string]bool {
	return flags
}

func GetFlag(ctx chain.QueryContext, name string) (bool, bool) {
	v, found := flags[name]
	return v, found
}

func GetUsers(ctx chain.QueryContext) map[string]Record {
	return users
}

func GetUser(ctx chain.QueryContext, name string) (Record, bool) {
	user, found := users[name]
	return user, found
}

func GetUserLimit(ctx chain.QueryContext, name string) (int64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}

	return user.Meta.Limit, true
}

func GetUserTagAt(ctx chain.QueryContext, name string, index int) (string, bool) {
	user, found := users[name]
	if !found {
		return "", false
	}
	if index < 0 || index >= len(user.Meta.Tags) {
		return "", false
	}

	return user.Meta.Tags[index], true
}

func GetWatchers(ctx chain.QueryContext) []Record {
	return watchers
}

func GetWatcherAt(ctx chain.QueryContext, index int) (Record, bool) {
	if index < 0 || index >= len(watchers) {
		return Record{}, false
	}

	return watchers[index], true
}

func GetWatcherLimitAt(ctx chain.QueryContext, index int) (int64, bool) {
	if index < 0 || index >= len(watchers) {
		return 0, false
	}

	return watchers[index].Meta.Limit, true
}

func GetCurrentContract(ctx chain.QueryContext) string {
	return ctx.GetContract()
}

func GetHeight(ctx chain.QueryContext) int64 {
	return ctx.GetHeight()
}

func GetCurrentHeight(ctx chain.QueryContext) int64 {
	return ctx.GetCurrentHeight()
}

func GetBlockTime(ctx chain.QueryContext) int64 {
	return blockTime
}

func DoesAccountExist(ctx chain.QueryContext, addr string) bool {
	return chain.AccountExists(addr)
}

func IsNamedContractAccount(ctx chain.QueryContext, addr string) bool {
	return chain.IsContractAccount(addr)
}

func GetBalanceOf(ctx chain.QueryContext, addr string, currency string) (string, bool) {
	return chain.BalanceOf(addr, currency)
}

func GetStdlibSummary(ctx chain.QueryContext) map[string]string {
	buf := bytes.NewBufferString(value)
	upper := strings.ToUpper(buf.String())
	hexValue := hex.EncodeToString([]byte(value))
	base64Value := base64.StdEncoding.EncodeToString([]byte(value))
	revisionText := strconv.FormatInt(revision, 10)
	utf8State := strconv.FormatBool(utf8.ValidString(value))

	return map[string]string{
		"upper":    upper,
		"hex":      hexValue,
		"base64":   base64Value,
		"revision": revisionText,
		"utf8":     utf8State,
	}
}

func ValidateUtf8Value(ctx chain.QueryContext) (string, bool) {
	if utf8.ValidString(value) {
		return value, true
	}

	return value, false
}

func GetValueIfPresent(ctx chain.QueryContext) (string, bool) {
	if value == "" {
		return "", false
	}

	return value, true
}
