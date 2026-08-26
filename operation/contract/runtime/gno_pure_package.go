package runtime

import (
	"embed"
	"fmt"
	"go/parser"
	"go/token"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
)

const (
	// These consensus-facing limits admit the pinned 9-file, 70 KiB uint256
	// package while leaving only modest room for reviewed pure dependencies.
	MaxGnoPurePackageCount       = 4
	MaxGnoPureDependencyDepth    = 8
	MaxGnoPureProductionBytes    = 128 * 1024
	MaxGnoPureFilesPerPackage    = 16
	gnoPurePackageEmbeddedPrefix = "gno_pure_packages"
)

//go:embed gno_pure_packages/gno.land/p/onbloc/uint256/*.gno gno_pure_packages/mitum/math/v1/u256/*.gno
var embeddedGnoPurePackages embed.FS

type GnoPurePackageLimits struct {
	MaxPackageCount    int
	MaxDependencyDepth int
	MaxSourceBytes     int
	MaxFilesPerPackage int
}

var defaultGnoPurePackageLimits = GnoPurePackageLimits{
	MaxPackageCount:    MaxGnoPurePackageCount,
	MaxDependencyDepth: MaxGnoPureDependencyDepth,
	MaxSourceBytes:     MaxGnoPureProductionBytes,
	MaxFilesPerPackage: MaxGnoPureFilesPerPackage,
}

var (
	gnoPurePackagesOnce   sync.Once
	gnoPurePackagesByRoot map[string][]*gnostd.MemPackage
	gnoPurePackagesErr    error
)

// GnoPureMemPackagesForContract returns fixed runtime packages only when the
// contract directly imports an exact canonical pure-package path.
func GnoPureMemPackagesForContract(sourceCode string) ([]*gnostd.MemPackage, error) {
	roots, err := contractImportsByKind(sourceCode, AllowedImportPurePackage)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, nil
	}
	if err := initializeGnoPurePackages(); err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	selected := make([]*gnostd.MemPackage, 0)
	for _, root := range roots {
		for _, pkg := range gnoPurePackagesByRoot[root] {
			if _, found := seen[pkg.Path]; found {
				continue
			}
			seen[pkg.Path] = struct{}{}
			selected = append(selected, pkg)
		}
	}
	return cloneGnoMemPackages(selected), nil
}

func initializeGnoPurePackages() error {
	gnoPurePackagesOnce.Do(func() {
		roots := allowedTypedContractImportPathsByKind(AllowedImportPurePackage)
		gnoPurePackagesByRoot = make(map[string][]*gnostd.MemPackage, len(roots))
		allowed := stringSet(embeddedGnoPurePackagePaths())
		for _, root := range roots {
			packages, err := loadGnoPureRuntimePackages(
				[]string{root}, allowed, readEmbeddedGnoPurePackage, defaultGnoPurePackageLimits,
			)
			if err != nil {
				gnoPurePackagesErr = err
				return
			}
			gnoPurePackagesByRoot[root] = packages
		}
	})
	return gnoPurePackagesErr
}

func loadGnoPureRuntimePackages(
	roots []string,
	allowed map[string]struct{},
	readPackage func(string) (*gnostd.MemPackage, error),
	limits GnoPurePackageLimits,
) ([]*gnostd.MemPackage, error) {
	purePackages, stdlibRoots, err := resolveGnoPurePackageGraph(roots, allowed, readPackage, limits)
	if err != nil {
		return nil, err
	}
	stdlibPackages, _, err := loadGnoStdlibMemPackages(stdlibRoots)
	if err != nil {
		return nil, fmt.Errorf("resolve pure package stdlib dependencies: %w", err)
	}
	return mergeGnoMemPackages(stdlibPackages, purePackages), nil
}

func resolveGnoPurePackageGraph(
	roots []string,
	allowed map[string]struct{},
	readPackage func(string) (*gnostd.MemPackage, error),
	limits GnoPurePackageLimits,
) ([]*gnostd.MemPackage, []string, error) {
	loaded := map[string]struct{}{}
	visiting := map[string]struct{}{}
	stdlibSet := map[string]struct{}{}
	packages := make([]*gnostd.MemPackage, 0, len(roots))
	totalBytes := 0

	orderedRoots := append([]string(nil), roots...)
	sort.Strings(orderedRoots)
	var visit func(string, int) error
	visit = func(importPath string, depth int) error {
		if _, found := loaded[importPath]; found {
			return nil
		}
		if _, found := visiting[importPath]; found {
			return fmt.Errorf("Gno pure package dependency cycle at %q", importPath)
		}
		if depth > limits.MaxDependencyDepth {
			return fmt.Errorf("Gno pure package dependency depth %d exceeds limit %d at %q", depth, limits.MaxDependencyDepth, importPath)
		}
		if len(loaded)+len(visiting) >= limits.MaxPackageCount {
			return fmt.Errorf("Gno pure package count exceeds limit %d at %q", limits.MaxPackageCount, importPath)
		}
		if _, found := allowed[importPath]; !found {
			return fmt.Errorf("Gno pure package %q is not in the exact-path policy", importPath)
		}

		visiting[importPath] = struct{}{}
		defer delete(visiting, importPath)
		pkg, err := readPackage(importPath)
		if err != nil {
			return err
		}
		if len(pkg.Files) > limits.MaxFilesPerPackage {
			return fmt.Errorf("Gno pure package %q has %d production files; limit is %d", importPath, len(pkg.Files), limits.MaxFilesPerPackage)
		}
		for _, file := range pkg.Files {
			totalBytes += len([]byte(file.Body))
			if totalBytes > limits.MaxSourceBytes {
				return fmt.Errorf("Gno pure package source bytes exceed limit %d at %q", limits.MaxSourceBytes, importPath)
			}
		}

		imports, err := gnoMemPackageImportPaths(pkg)
		if err != nil {
			return err
		}
		for _, dep := range imports {
			if _, found := allowed[dep]; found {
				if err := visit(dep, depth+1); err != nil {
					return err
				}
				continue
			}
			if gno.IsStdlib(dep) {
				stdlibSet[dep] = struct{}{}
				continue
			}
			return fmt.Errorf("Gno pure package %q imports unsupported package %q", importPath, dep)
		}
		loaded[importPath] = struct{}{}
		packages = append(packages, pkg)
		return nil
	}

	for _, root := range orderedRoots {
		if err := visit(root, 1); err != nil {
			return nil, nil, err
		}
	}
	stdlibRoots := make([]string, 0, len(stdlibSet))
	for importPath := range stdlibSet {
		stdlibRoots = append(stdlibRoots, importPath)
	}
	sort.Strings(stdlibRoots)
	return packages, stdlibRoots, nil
}

func readEmbeddedGnoPurePackage(importPath string) (*gnostd.MemPackage, error) {
	if _, found := stringSet(embeddedGnoPurePackagePaths())[importPath]; !found {
		return nil, fmt.Errorf("Gno pure package %q is not embedded", importPath)
	}
	dir := path.Join(gnoPurePackageEmbeddedPrefix, importPath)
	entries, err := embeddedGnoPurePackages.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read embedded Gno pure package %q: %w", importPath, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".gno") ||
			strings.HasSuffix(name, "_test.gno") || strings.HasSuffix(name, "_filetest.gno") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("embedded Gno pure package %q contains no production .gno files", importPath)
	}
	files := make([]*gnostd.MemFile, 0, len(names))
	for _, name := range names {
		body, err := embeddedGnoPurePackages.ReadFile(path.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read embedded Gno pure file %q/%q: %w", importPath, name, err)
		}
		files = append(files, &gnostd.MemFile{Name: name, Body: string(body)})
	}
	packageType := gno.MPUserProd
	if importPath == Uint256PackagePath {
		// GnoVM restricts MPUserProd identities to gno.land user paths. The
		// system-owned hybrid API remains exact-policy code but must be stored
		// as MPStdlibProd to preserve its required mitum/math/... identity.
		packageType = gno.MPStdlibProd
	}
	pkg := &gnostd.MemPackage{Name: path.Base(importPath), Path: importPath, Type: packageType, Files: files}
	if err := validateGnoPurePackageIdentity(pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

func validateGnoPurePackageIdentity(pkg *gnostd.MemPackage) error {
	expected := path.Base(pkg.Path)
	if pkg.Name != expected {
		return fmt.Errorf("Gno pure package %q name %q does not match path basename %q", pkg.Path, pkg.Name, expected)
	}
	for _, file := range pkg.Files {
		node, err := parser.ParseFile(token.NewFileSet(), file.Name, file.Body, parser.PackageClauseOnly)
		if err != nil {
			return fmt.Errorf("parse Gno package declaration in %q/%q: %w", pkg.Path, file.Name, err)
		}
		if node.Name.Name != expected {
			return fmt.Errorf("Gno pure package %q file %q declares package %q; expected %q", pkg.Path, file.Name, node.Name.Name, expected)
		}
	}
	return nil
}

func embeddedGnoPurePackagePaths() []string {
	return []string{OnblocUint256PackagePath, Uint256PackagePath}
}

func contractImportsByKind(sourceCode string, kind AllowedImportKind) ([]string, error) {
	node, err := parser.ParseFile(token.NewFileSet(), "contract.gno", sourceCode, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse contract imports: %w", err)
	}
	roots := allowedTypedContractImportPathsByKind(kind)
	allowed := stringSet(roots)
	imported := map[string]struct{}{}
	for _, imp := range node.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("decode contract import: %w", err)
		}
		if _, found := allowed[importPath]; found {
			imported[importPath] = struct{}{}
		}
	}
	out := make([]string, 0, len(imported))
	for _, root := range roots {
		if _, found := imported[root]; found {
			out = append(out, root)
		}
	}
	return out, nil
}

func gnoMemPackageImportPaths(pkg *gnostd.MemPackage) ([]string, error) {
	seen := map[string]struct{}{}
	for _, file := range pkg.Files {
		node, err := parser.ParseFile(token.NewFileSet(), file.Name, file.Body, parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("parse Gno imports in %q/%q: %w", pkg.Path, file.Name, err)
		}
		for _, imp := range node.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return nil, fmt.Errorf("decode Gno import in %q/%q: %w", pkg.Path, file.Name, err)
			}
			seen[importPath] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for importPath := range seen {
		out = append(out, importPath)
	}
	sort.Strings(out)
	return out, nil
}

func mergeGnoMemPackages(groups ...[]*gnostd.MemPackage) []*gnostd.MemPackage {
	seen := map[string]struct{}{}
	out := make([]*gnostd.MemPackage, 0)
	for _, packages := range groups {
		for _, pkg := range packages {
			if _, found := seen[pkg.Path]; found {
				continue
			}
			seen[pkg.Path] = struct{}{}
			out = append(out, pkg)
		}
	}
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}
