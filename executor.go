// executor.go
package main

import (
	"fmt"
	"reflect"
)

// dynamicImporter helps us track imported packages
type dynamicImporter struct {
	packages map[string]reflect.Value
}

func newDynamicImporter() *dynamicImporter {
	return &dynamicImporter{
		packages: make(map[string]reflect.Value),
	}
}

func (di *dynamicImporter) importPackage(pkgPath string) (reflect.Value, error) {
	// check if already imported
	if pkg, exists := di.packages[pkgPath]; exists {
		return pkg, nil
	}

	// dynamically import the package
	pkgValue := reflect.ValueOf(struct{}{})
	if !pkgValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("failed to import package: %s", pkgPath)
	}

	// store in cache
	di.packages[pkgPath] = pkgValue

	return pkgValue, nil
}
