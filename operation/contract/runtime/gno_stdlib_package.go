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

//go:embed gno_stdlib
var embeddedGnoStdlib embed.FS

var (
	gnoStdlibPackagesOnce        sync.Once
	gnoStdlibPackages            []*gnostd.MemPackage
	gnoStdlibPackageDependencies map[string][]string
	gnoStdlibPackagesErr         error
)

// GnoStdlibMemPackages returns the Gno source packages needed by the allowed
// contract stdlib imports. The Gno source is vendored under gno_stdlib and
// embedded into the binary, so runtime execution does not depend on GNOROOT,
// GOMODCACHE, or a local github.com/gnolang/gno checkout.
//
// Internal dependencies are loaded to execute allowed packages, but are not
// added to the contract import allowlist.
func GnoStdlibMemPackages() ([]*gnostd.MemPackage, error) {
	if err := initializeGnoStdlibPackages(); err != nil {
		return nil, err
	}

	return cloneGnoMemPackages(gnoStdlibPackages), nil
}

// GnoStdlibMemPackagesForContract returns only directly imported allowed
// stdlib packages and their internal dependencies. This keeps contracts which
// import no stdlib from paying package initialization cost for unused code.
func GnoStdlibMemPackagesForContract(sourceCode string) ([]*gnostd.MemPackage, error) {
	roots, err := contractGnoStdlibImports(sourceCode)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, nil
	}

	if err := initializeGnoStdlibPackages(); err != nil {
		return nil, err
	}

	included := map[string]bool{}
	var include func(string)
	include = func(importPath string) {
		if included[importPath] {
			return
		}
		for _, dep := range gnoStdlibPackageDependencies[importPath] {
			include(dep)
		}
		included[importPath] = true
	}
	for _, importPath := range roots {
		include(importPath)
	}

	selected := make([]*gnostd.MemPackage, 0, len(included))
	for _, pkg := range gnoStdlibPackages {
		if included[pkg.Path] {
			selected = append(selected, pkg)
		}
	}

	return cloneGnoMemPackages(selected), nil
}

func initializeGnoStdlibPackages() error {
	gnoStdlibPackagesOnce.Do(func() {
		gnoStdlibPackages, gnoStdlibPackageDependencies, gnoStdlibPackagesErr = loadGnoStdlibMemPackages(allowedTypedContractImportPathsByKind(AllowedImportStdlib))
	})

	return gnoStdlibPackagesErr
}

func loadGnoStdlibMemPackages(roots []string) ([]*gnostd.MemPackage, map[string][]string, error) {
	loaded := map[string]bool{}
	visiting := map[string]bool{}
	dependencies := map[string][]string{}
	packages := make([]*gnostd.MemPackage, 0, len(roots))

	var visit func(string) error
	visit = func(importPath string) error {
		if loaded[importPath] {
			return nil
		}
		if visiting[importPath] {
			return fmt.Errorf("gno stdlib dependency cycle at %q", importPath)
		}
		visiting[importPath] = true
		defer delete(visiting, importPath)

		pkg, err := readGnoStdlibMemPackage(importPath)
		if err != nil {
			return err
		}

		deps, err := gnoStdlibImportPaths(pkg)
		if err != nil {
			return err
		}
		dependencies[importPath] = deps
		for _, dep := range deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		loaded[importPath] = true
		packages = append(packages, pkg)
		return nil
	}

	for _, importPath := range roots {
		if err := visit(importPath); err != nil {
			return nil, nil, err
		}
	}

	return packages, dependencies, nil
}

func contractGnoStdlibImports(sourceCode string) ([]string, error) {
	node, err := parser.ParseFile(token.NewFileSet(), "contract.gno", sourceCode, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse contract stdlib imports: %w", err)
	}

	roots := allowedTypedContractImportPathsByKind(AllowedImportStdlib)
	rootSet := make(map[string]struct{}, len(roots))
	for _, importPath := range roots {
		rootSet[importPath] = struct{}{}
	}

	imported := map[string]struct{}{}
	for _, imp := range node.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("decode contract stdlib import: %w", err)
		}
		if _, found := rootSet[importPath]; found {
			imported[importPath] = struct{}{}
		}
	}

	out := make([]string, 0, len(imported))
	for _, importPath := range roots {
		if _, found := imported[importPath]; found {
			out = append(out, importPath)
		}
	}
	return out, nil
}

func readGnoStdlibMemPackage(importPath string) (*gnostd.MemPackage, error) {
	dir := path.Join("gno_stdlib", importPath)
	entries, err := embeddedGnoStdlib.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read Gno stdlib package %q: %w", importPath, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".gno") ||
			strings.HasSuffix(name, "_test.gno") ||
			strings.HasSuffix(name, "_filetest.gno") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		return nil, fmt.Errorf("Gno stdlib package %q contains no production .gno files", importPath)
	}

	files := make([]*gnostd.MemFile, 0, len(names))
	for _, name := range names {
		body, err := embeddedGnoStdlib.ReadFile(path.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read Gno stdlib file %q/%q: %w", importPath, name, err)
		}
		files = append(files, &gnostd.MemFile{Name: name, Body: string(body)})
	}

	return &gnostd.MemPackage{
		Name:  path.Base(importPath),
		Path:  importPath,
		Type:  gno.MPStdlibProd,
		Files: files,
	}, nil
}

func gnoStdlibImportPaths(pkg *gnostd.MemPackage) ([]string, error) {
	seen := map[string]struct{}{}

	for _, file := range pkg.Files {
		node, err := parser.ParseFile(token.NewFileSet(), file.Name, file.Body, parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("parse Gno stdlib imports in %q/%q: %w", pkg.Path, file.Name, err)
		}

		for _, imp := range node.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return nil, fmt.Errorf("decode Gno stdlib import in %q/%q: %w", pkg.Path, file.Name, err)
			}
			if !gno.IsStdlib(importPath) {
				return nil, fmt.Errorf("Gno stdlib package %q imports non-stdlib package %q", pkg.Path, importPath)
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

func cloneGnoMemPackages(packages []*gnostd.MemPackage) []*gnostd.MemPackage {
	out := make([]*gnostd.MemPackage, 0, len(packages))

	for _, pkg := range packages {
		files := make([]*gnostd.MemFile, 0, len(pkg.Files))
		for _, file := range pkg.Files {
			files = append(files, file.Copy())
		}
		out = append(out, &gnostd.MemPackage{
			Name:  pkg.Name,
			Path:  pkg.Path,
			Type:  pkg.Type,
			Info:  pkg.Info,
			Files: files,
		})
	}

	return out
}
