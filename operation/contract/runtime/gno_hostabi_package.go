package runtime

import gnostd "github.com/gnolang/gno/tm2/pkg/std"

const (
	MitumChainPackagePath = "mitum/chain"
	MitumChainPackageName = "chain"
)

const mitumChainPackageSource = `package chain

type ContractContext struct {
	Sender   string
	Contract string
	Height   int64
	ReadOnly bool
}

func (ctx ContractContext) GetSender() string {
	return ctx.Sender
}

func (ctx ContractContext) GetContract() string {
	return ctx.Contract
}

func (ctx ContractContext) GetHeight() int64 {
	return ctx.Height
}

func (ctx ContractContext) IsReadOnly() bool {
	return ctx.ReadOnly
}

func AccountExists(addr string) bool
func IsContractAccount(addr string) bool
`

func MitumChainMemPackage() *gnostd.MemPackage {
	return &gnostd.MemPackage{
		Name: MitumChainPackageName,
		Path: MitumChainPackagePath,
		Files: []*gnostd.MemFile{
			{
				Name: "chain.gno",
				Body: mitumChainPackageSource,
			},
		},
	}
}

func HostABIMemPackages() []*gnostd.MemPackage {
	return []*gnostd.MemPackage{
		MitumChainMemPackage(),
	}
}
