//go:build ignore
// +build ignore

// Runtime-only Gno contract example. This file is not part of the normal Go build.
package contract

import "mitum/chain"

type contractError string

func (e contractError) Error() string {
	return string(e)
}

type Limits struct {
	Daily int64
	Max   uint64
}

type Profile struct {
	Active   bool
	Nickname string
	Level    int
}

type UserSelector struct {
	Name          string
	RequireActive bool
}

type UserMeta struct {
	Active bool
	Limit  int64
	Tags   []string
	Labels map[string]string
}

type User struct {
	Name    string
	Balance int64
	Counter uint64
	Meta    UserMeta
	Flags   map[string]bool
	Scores  []int64
}

type Config struct {
	Enabled       bool
	Version       int
	Limits        Limits
	Primary       UserMeta
	FeatureFlags  map[string]bool
	FeatureLabels map[string]string
	Aliases       []string
	Members       map[string]User
	Reviewers     []User
}

var initialized bool
var owner string
var value string
var revision int64
var generation int
var supply uint64
var limits Limits
var profile Profile
var flags map[string]bool
var labels map[string]string
var users map[string]User
var admins []string
var scores []int64
var watchers []User
var config Config

func Initialize(ctx chain.ContractContext) error {
	if initialized {
		return nil
	}

	owner = ctx.GetSender()
	value = ""
	revision = 0
	generation = 7
	supply = 1000

	limits = Limits{
		Daily: 100,
		Max:   1000,
	}

	profile = Profile{
		Active:   true,
		Nickname: "owner-profile",
		Level:    3,
	}

	flags = map[string]bool{
		"alpha": true,
		"beta":  false,
	}

	labels = map[string]string{
		"network": "test",
		"region":  "seoul",
	}

	users = map[string]User{
		"alice": User{
			Name:    "alice",
			Balance: 10,
			Counter: 1,
			Meta: UserMeta{
				Active: true,
				Limit:  50,
				Tags:   []string{"tag-a", "tag-b"},
				Labels: map[string]string{
					"tier": "gold",
				},
			},
			Flags: map[string]bool{
				"verified": true,
			},
			Scores: []int64{10, 20},
		},
		"bob": User{
			Name:    "bob",
			Balance: 5,
			Counter: 2,
			Meta: UserMeta{
				Active: false,
				Limit:  25,
				Tags:   []string{"tag-c"},
				Labels: map[string]string{
					"tier": "silver",
				},
			},
			Flags: map[string]bool{
				"verified": false,
			},
			Scores: []int64{7},
		},
	}

	admins = []string{"alice", "bob"}
	scores = []int64{3, 5, 8}

	watchers = []User{
		User{
			Name:    "watcher-a",
			Balance: 11,
			Counter: 1,
			Meta: UserMeta{
				Active: true,
				Limit:  90,
				Tags:   []string{"watch"},
				Labels: map[string]string{
					"role": "observer",
				},
			},
			Flags: map[string]bool{
				"primary": true,
			},
			Scores: []int64{100},
		},
		User{
			Name:    "watcher-b",
			Balance: 13,
			Counter: 2,
			Meta: UserMeta{
				Active: false,
				Limit:  70,
				Tags:   []string{"backup"},
				Labels: map[string]string{
					"role": "standby",
				},
			},
			Flags: map[string]bool{
				"primary": false,
			},
			Scores: []int64{80, 81},
		},
	}

	config = Config{
		Enabled: true,
		Version: 2,
		Limits: Limits{
			Daily: 250,
			Max:   5000,
		},
		Primary: UserMeta{
			Active: true,
			Limit:  120,
			Tags:   []string{"primary-a", "primary-b"},
			Labels: map[string]string{
				"scope": "global",
			},
		},
		FeatureFlags: map[string]bool{
			"search": true,
			"write":  false,
		},
		FeatureLabels: map[string]string{
			"env":   "dev",
			"owner": "ops",
		},
		Aliases: []string{"core", "backup"},
		Members: map[string]User{
			"charlie": User{
				Name:    "charlie",
				Balance: 30,
				Counter: 3,
				Meta: UserMeta{
					Active: true,
					Limit:  200,
					Tags:   []string{"member"},
					Labels: map[string]string{
						"group": "core",
					},
				},
				Flags: map[string]bool{
					"admin": true,
				},
				Scores: []int64{30, 31},
			},
		},
		Reviewers: []User{
			User{
				Name:    "reviewer-a",
				Balance: 41,
				Counter: 1,
				Meta: UserMeta{
					Active: true,
					Limit:  300,
					Tags:   []string{"rv-a"},
					Labels: map[string]string{
						"shift": "day",
					},
				},
				Flags: map[string]bool{
					"lead": true,
				},
				Scores: []int64{1, 2},
			},
		},
	}

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

	value = data
	revision = 1

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

	value = data
	revision = revision + 1

	return nil
}

func SetLimits(ctx chain.ContractContext, next Limits) error {
	limits = next
	revision = revision + 1

	return nil
}

func SetProfile(ctx chain.ContractContext, next Profile) error {
	profile = next
	revision = revision + 1

	return nil
}

func SetFlags(ctx chain.ContractContext, next map[string]bool) error {
	flags = next
	revision = revision + 1

	return nil
}

func SetLabels(ctx chain.ContractContext, next map[string]string) error {
	labels = next
	revision = revision + 1

	return nil
}

func ReplaceUsers(ctx chain.ContractContext, next map[string]User) error {
	users = next
	revision = revision + 1

	return nil
}

func SetAdmins(ctx chain.ContractContext, next []string) error {
	admins = next
	revision = revision + 1

	return nil
}

func SetScores(ctx chain.ContractContext, next []int64) error {
	scores = next
	revision = revision + 1

	return nil
}

func ReplaceWatchers(ctx chain.ContractContext, next []User) error {
	watchers = next
	revision = revision + 1

	return nil
}

func SetConfig(ctx chain.ContractContext, next Config) error {
	config = next
	revision = revision + 1

	return nil
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

func GetGeneration(ctx chain.ContractContext) int {
	return generation
}

func GetSupply(ctx chain.ContractContext) uint64 {
	return supply
}

func GetDailyLimit(ctx chain.ContractContext) int64 {
	return limits.Daily
}

func GetMaxLimit(ctx chain.ContractContext) uint64 {
	return limits.Max
}

func IsProfileActive(ctx chain.ContractContext) bool {
	return profile.Active
}

func GetProfileNickname(ctx chain.ContractContext) string {
	return profile.Nickname
}

func GetProfileLevel(ctx chain.ContractContext) int {
	return profile.Level
}

func GetFlag(ctx chain.ContractContext, name string) (bool, bool) {
	v, found := flags[name]
	return v, found
}

func GetLabel(ctx chain.ContractContext, name string) (string, bool) {
	v, found := labels[name]
	return v, found
}

func GetLimits(ctx chain.ContractContext) Limits {
	return limits
}

func GetProfile(ctx chain.ContractContext) Profile {
	return profile
}

func GetFlags(ctx chain.ContractContext) map[string]bool {
	return flags
}

func GetLabels(ctx chain.ContractContext) map[string]string {
	return labels
}

func GetUsers(ctx chain.ContractContext) map[string]User {
	return users
}

func GetAdmins(ctx chain.ContractContext) []string {
	return admins
}

func GetScores(ctx chain.ContractContext) []int64 {
	return scores
}

func GetWatchers(ctx chain.ContractContext) []User {
	return watchers
}

func GetConfig(ctx chain.ContractContext) Config {
	return config
}

func GetSelectedUser(ctx chain.ContractContext, selector UserSelector) (User, bool) {
	user, found := users[selector.Name]
	if !found {
		return User{}, false
	}
	if selector.RequireActive && !user.Meta.Active {
		return User{}, false
	}

	return user, true
}

func EchoFlags(ctx chain.ContractContext, next map[string]bool) map[string]bool {
	return next
}

func EchoUsers(ctx chain.ContractContext, next map[string]User) map[string]User {
	return next
}

func EchoAdmins(ctx chain.ContractContext, next []string) []string {
	return next
}

func EchoWatchers(ctx chain.ContractContext, next []User) []User {
	return next
}

func GetUserName(ctx chain.ContractContext, name string) (string, bool) {
	user, found := users[name]
	if !found {
		return "", false
	}
	return user.Name, true
}

func GetUserBalance(ctx chain.ContractContext, name string) (int64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}
	return user.Balance, true
}

func GetUserCounter(ctx chain.ContractContext, name string) (uint64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}
	return user.Counter, true
}

func GetUserMetaActive(ctx chain.ContractContext, name string) (bool, bool) {
	user, found := users[name]
	if !found {
		return false, false
	}
	return user.Meta.Active, true
}

func GetUserMetaLimit(ctx chain.ContractContext, name string) (int64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}
	return user.Meta.Limit, true
}

func GetUserMetaTagAt(ctx chain.ContractContext, name string, index int) (string, bool) {
	user, found := users[name]
	if !found {
		return "", false
	}
	if index < 0 || index >= len(user.Meta.Tags) {
		return "", false
	}
	return user.Meta.Tags[index], true
}

func GetUserMetaLabel(ctx chain.ContractContext, name string, key string) (string, bool) {
	user, found := users[name]
	if !found {
		return "", false
	}
	v, ok := user.Meta.Labels[key]
	return v, ok
}

func GetUserFlag(ctx chain.ContractContext, name string, key string) (bool, bool) {
	user, found := users[name]
	if !found {
		return false, false
	}
	v, ok := user.Flags[key]
	return v, ok
}

func GetUserScoreAt(ctx chain.ContractContext, name string, index int) (int64, bool) {
	user, found := users[name]
	if !found {
		return 0, false
	}
	if index < 0 || index >= len(user.Scores) {
		return 0, false
	}
	return user.Scores[index], true
}

func GetAdminsCount(ctx chain.ContractContext) int64 {
	return int64(len(admins))
}

func GetAdminAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(admins) {
		return "", false
	}
	return admins[index], true
}

func GetScoresCount(ctx chain.ContractContext) int64 {
	return int64(len(scores))
}

func GetScoreAt(ctx chain.ContractContext, index int) (int64, bool) {
	if index < 0 || index >= len(scores) {
		return 0, false
	}
	return scores[index], true
}

func GetWatchersCount(ctx chain.ContractContext) int64 {
	return int64(len(watchers))
}

func GetWatcherNameAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(watchers) {
		return "", false
	}
	return watchers[index].Name, true
}

func GetWatcherLimitAt(ctx chain.ContractContext, index int) (int64, bool) {
	if index < 0 || index >= len(watchers) {
		return 0, false
	}
	return watchers[index].Meta.Limit, true
}

func IsConfigEnabled(ctx chain.ContractContext) bool {
	return config.Enabled
}

func GetConfigVersion(ctx chain.ContractContext) int {
	return config.Version
}

func GetConfigLimitDaily(ctx chain.ContractContext) int64 {
	return config.Limits.Daily
}

func GetConfigLimitMax(ctx chain.ContractContext) uint64 {
	return config.Limits.Max
}

func GetConfigPrimaryActive(ctx chain.ContractContext) bool {
	return config.Primary.Active
}

func GetConfigPrimaryLimit(ctx chain.ContractContext) int64 {
	return config.Primary.Limit
}

func GetConfigPrimaryTagAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(config.Primary.Tags) {
		return "", false
	}
	return config.Primary.Tags[index], true
}

func GetConfigPrimaryLabel(ctx chain.ContractContext, key string) (string, bool) {
	v, ok := config.Primary.Labels[key]
	return v, ok
}

func GetConfigFeatureFlag(ctx chain.ContractContext, key string) (bool, bool) {
	v, ok := config.FeatureFlags[key]
	return v, ok
}

func GetConfigFeatureLabel(ctx chain.ContractContext, key string) (string, bool) {
	v, ok := config.FeatureLabels[key]
	return v, ok
}

func GetConfigAliasesCount(ctx chain.ContractContext) int64 {
	return int64(len(config.Aliases))
}

func GetConfigAliasAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(config.Aliases) {
		return "", false
	}
	return config.Aliases[index], true
}

func GetConfigMemberBalance(ctx chain.ContractContext, name string) (int64, bool) {
	member, found := config.Members[name]
	if !found {
		return 0, false
	}
	return member.Balance, true
}

func GetConfigMemberMetaLimit(ctx chain.ContractContext, name string) (int64, bool) {
	member, found := config.Members[name]
	if !found {
		return 0, false
	}
	return member.Meta.Limit, true
}

func GetConfigReviewersCount(ctx chain.ContractContext) int64 {
	return int64(len(config.Reviewers))
}

func GetConfigReviewerNameAt(ctx chain.ContractContext, index int) (string, bool) {
	if index < 0 || index >= len(config.Reviewers) {
		return "", false
	}
	return config.Reviewers[index].Name, true
}

func GetValueIfPresent(ctx chain.ContractContext) (string, bool) {
	if value == "" {
		return "", false
	}

	return value, true
}
