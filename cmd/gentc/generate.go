package main

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"

	"github.com/sijad/gentle/builder"
	"golang.org/x/tools/go/packages"
)

func generate(schemaPath string, generatedPath string) error {
	pkgPath, err := packagePath(schemaPath)

	if err != nil {
		return err
	}

	cfg := packages.Config{Mode: packages.NeedTypesInfo | packages.NeedTypes}
	pkgs, err := packages.Load(&cfg, pkgPath)

	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	query, err := lookupTypeName("Query", pkgs)
	if err != nil {
		return fmt.Errorf("generate code Object lookup %s: %w", "Query", err)
	}

	mutation, err := lookupTypeName("Mutation", pkgs)
	if err != nil {
		return fmt.Errorf("generate code Object lookup %s: %w", "Mutation", err)
	}

	if query == nil && mutation == nil {
		return fmt.Errorf("generate code: Query or Mutation structs not found")
	}

	bldr := builder.NewGQLBuilder()

	if query != nil {
		if _, err := bldr.ImportType(query.Type()); err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
	}

	if mutation != nil {
		if _, err := bldr.ImportType(mutation.Type()); err != nil {
			return fmt.Errorf("generate code: %w", err)
		}
	}

	file, err := os.Create(generatedPath)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}
	defer file.Close()

	if err := bldr.Code(file); err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	return nil
}

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

	var parts []string

	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		parts = append(parts, filepath.Base(realPath))
		realPath = filepath.Dir(realPath)
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedName}, realPath)
	if err != nil {
		return "", fmt.Errorf("load package info: %v", err)
	}
	pkgPath := pkgs[0].PkgPath

	return pkgPath, nil
}
