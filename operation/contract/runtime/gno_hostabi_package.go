package runtime

import gnostd "github.com/gnolang/gno/tm2/pkg/std"

const (
	MitumChainPackageName = "chain"
)

const mitumChainPackageSource = `package chain

type WriteContext struct {
	Sender    string
	Contract  string
	Height    int64
	BlockTime int64
	ReadOnly  bool
}

func (ctx WriteContext) GetSender() string {
	return ctx.Sender
}

func (ctx WriteContext) GetContract() string {
	return ctx.Contract
}

func (ctx WriteContext) GetHeight() int64 {
	return ctx.Height
}

func (ctx WriteContext) GetBlockTime() int64 {
	return ctx.BlockTime
}

func (ctx WriteContext) IsReadOnly() bool {
	return ctx.ReadOnly
}

type QueryContext struct {
	Contract      string
	Height        int64
	CurrentHeight int64
	ReadOnly      bool
}

func (ctx QueryContext) GetContract() string {
	return ctx.Contract
}

func (ctx QueryContext) GetHeight() int64 {
	return ctx.Height
}

func (ctx QueryContext) GetCurrentHeight() int64 {
	return ctx.CurrentHeight
}

func (ctx QueryContext) IsReadOnly() bool {
	return ctx.ReadOnly
}

func AccountExists(addr string) bool
func IsContractAccount(addr string) bool
func BalanceOf(addr string, currency string) (string, bool)
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
	paths := allowedTypedContractImportPathsByKind(AllowedImportHostABI)
	packages := make([]*gnostd.MemPackage, 0, len(paths))
	for _, importPath := range paths {
		switch importPath {
		case MitumChainPackagePath:
			packages = append(packages, MitumChainMemPackage())
		default:
			panic("canonical host ABI import has no MemPackage implementation: " + importPath)
		}
	}
	return packages
}
