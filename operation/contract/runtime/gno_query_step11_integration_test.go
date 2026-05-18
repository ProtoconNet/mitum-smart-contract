package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const scalarQueryArgContractSource = `package contract
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
	Users   map[string]User
	Aliases []string
}

var config Config

func Initialize(ctx chain.ContractContext) error {
	config.Users = map[string]User{
		"alice": User{Name: "alice", Balance: 10, Meta: Meta{Active: true, Limit: 100}},
		"bob": User{Name: "bob", Balance: 20, Meta: Meta{Active: false, Limit: 200}},
	}
	config.Aliases = []string{"root", "child"}
	return nil
}

func GetSelectedUser(ctx chain.ContractContext, name string, requireActive bool) (User, bool) {
	user, found := config.Users[name]
	if !found {
		return User{}, false
	}
	if requireActive && !user.Meta.Active {
		return User{}, false
	}
	return user, true
}

func HasAlias(ctx chain.ContractContext, name string) bool {
	for _, alias := range config.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}
`

func TestGnoQueryPathScalarArgsStillSupportCompositeResult(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		scalarQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(400),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: scalarQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, scalarQueryArgContractSource, "GetSelectedUser", map[string]string{
		"name":          "alice",
		"requireActive": "true",
	})
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetSelectedUser ok=true, got %#v", qr.Ok)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Name":    "alice",
		"Balance": int64(10),
		"Meta": map[string]interface{}{
			"Active": true,
			"Limit":  int64(100),
		},
	})

	qr = mustQueryContract(t, states, contract, sender, scalarQueryArgContractSource, "GetSelectedUser", map[string]string{
		"name":          "bob",
		"requireActive": "true",
	})
	if qr.Ok == nil || *qr.Ok {
		t.Fatalf("expected GetSelectedUser ok=false, got %#v", qr.Ok)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Name":    "",
		"Balance": int64(0),
		"Meta": map[string]interface{}{
			"Active": false,
			"Limit":  int64(0),
		},
	})

	qr = mustQueryContract(t, states, contract, sender, scalarQueryArgContractSource, "HasAlias", map[string]string{
		"name": "root",
	})
	if got := qr.Result.(bool); !got {
		t.Fatalf("expected HasAlias(root)=true")
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
