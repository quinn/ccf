package content

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
)

func loadPackage(dir string) (*types.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	if len(pkgs[0].Errors) > 0 {
		return nil, fmt.Errorf("package has errors: %v", pkgs[0].Errors)
	}

	return pkgs[0].Types, nil
}
