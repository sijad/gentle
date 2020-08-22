package builder

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func lookupTypeName(name string, pkgs []*packages.Package) (typ *types.TypeName, err error) {
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(name)

		if obj != nil {
			if typ != nil {
				return nil, errors.New("multiple Object with same name found")
			}
			t, ok := obj.(*types.TypeName)
			if !ok {
				return nil, errors.New("type must be a Defined or Alias")
			}
			if _, ok := t.Type().(*types.Named); !ok {
				return nil, errors.New("type must be a named struct")
			}
			typ = t
		}
	}

	return
}

func packagePath(target string) (string, error) {
	realPath, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		realPath = filepath.Dir(realPath)
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, realPath)
	if err != nil {
		return "", fmt.Errorf("load package info: %v", err)
	}
	pkgPath := pkgs[0].PkgPath

	return pkgPath, nil
}

func loadPackages(pkgPath string) ([]*packages.Package, error) {
	cfg := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(&cfg, pkgPath)

	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
