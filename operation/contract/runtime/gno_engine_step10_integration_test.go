package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const scalarWriteArgContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Active bool
	Limit  int64
}

type User struct {
	Name    string
	Balance int64
	Meta    Meta
}

type Config struct {
	Label   string
	Flags   map[string]bool
	Users   map[string]User
	Aliases []string
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Label = "init"
	config.Flags = map[string]bool{"alpha": true}
	config.Users = map[string]User{
		"alice": User{Name: "alice", Balance: 10, Meta: Meta{Active: true, Limit: 100}},
	}
	config.Aliases = []string{"root"}
	return nil
}

func SetLabel(ctx chain.WriteContext, next string) error {
	config.Label = next
	return nil
}

func SetFlag(ctx chain.WriteContext, name string, enabled bool) error {
	if config.Flags == nil {
		config.Flags = map[string]bool{}
	}
	config.Flags[name] = enabled
	return nil
}

func AddAlias(ctx chain.WriteContext, alias string) error {
	config.Aliases = append(config.Aliases, alias)
	return nil
}

func SetUserLimit(ctx chain.WriteContext, name string, limit int64) error {
	user, found := config.Users[name]
	if !found {
		return nil
	}
	user.Meta.Limit = limit
	config.Users[name] = user
	return nil
}

func GetConfig(ctx chain.QueryContext) Config { return config }
func GetLabel(ctx chain.QueryContext) string { return config.Label }
`

func TestGnoWritePathScalarArgsStillMutateCompositeState(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		scalarWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(300),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: scalarWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(301),
				Function:     "SetLabel",
				CallData:     map[string]string{"next": "live"},
				ContractCode: scalarWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(302),
				Function:     "SetFlag",
				CallData:     map[string]string{"name": "beta", "enabled": "true"},
				ContractCode: scalarWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(303),
				Function:     "AddAlias",
				CallData:     map[string]string{"alias": "child"},
				ContractCode: scalarWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(304),
				Function:     "SetUserLimit",
				CallData:     map[string]string{"name": "alice", "limit": "777"},
				ContractCode: scalarWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, scalarWriteArgContractSource, "GetConfig", map[string]string{})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Label": "live",
		"Flags": map[string]interface{}{
			"alpha": true,
			"beta":  true,
		},
		"Users": map[string]interface{}{
			"alice": map[string]interface{}{
				"Name":    "alice",
				"Balance": int64(10),
				"Meta": map[string]interface{}{
					"Active": true,
					"Limit":  int64(777),
				},
			},
		},
		"Aliases": []interface{}{"root", "child"},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
