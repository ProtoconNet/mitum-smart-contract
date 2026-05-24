package runtime

const MitumChainPackagePath = "mitum/chain"

type AllowedImportKind string

const (
	AllowedImportHostABI AllowedImportKind = "hostabi"
	AllowedImportStdlib  AllowedImportKind = "stdlib"
)

type AllowedImportSpec struct {
	Path string
	Kind AllowedImportKind
}

var allowedTypedContractImportSpecs = [...]AllowedImportSpec{
	{Path: MitumChainPackagePath, Kind: AllowedImportHostABI},
	{Path: "strconv", Kind: AllowedImportStdlib},
	{Path: "strings", Kind: AllowedImportStdlib},
	{Path: "errors", Kind: AllowedImportStdlib},
	{Path: "bytes", Kind: AllowedImportStdlib},
	{Path: "encoding/hex", Kind: AllowedImportStdlib},
	{Path: "encoding/base64", Kind: AllowedImportStdlib},
	{Path: "unicode/utf8", Kind: AllowedImportStdlib},
}

// AllowedTypedContractImportSpecs returns the canonical contract-facing import
// policy. Schema admission and runtime package registration must derive from
// this list so an admitted import cannot silently lose runtime support.
func AllowedTypedContractImportSpecs() []AllowedImportSpec {
	out := make([]AllowedImportSpec, len(allowedTypedContractImportSpecs))
	copy(out, allowedTypedContractImportSpecs[:])
	return out
}

func allowedTypedContractImportPaths() []string {
	specs := AllowedTypedContractImportSpecs()
	out := make([]string, 0, len(specs))
	for _, spec := range specs {
		out = append(out, spec.Path)
	}
	return out
}

func allowedTypedContractImportPathsByKind(kind AllowedImportKind) []string {
	specs := AllowedTypedContractImportSpecs()
	out := make([]string, 0, len(specs))
	for _, spec := range specs {
		if spec.Kind == kind {
			out = append(out, spec.Path)
		}
	}
	return out
}

func allowedTypedContractImportPathSet() map[string]struct{} {
	specs := AllowedTypedContractImportSpecs()
	out := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		out[spec.Path] = struct{}{}
	}
	return out
}
