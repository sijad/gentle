package main

import (
	"fmt"
	"os"

	"github.com/sijad/gentle/builder"
)

func generate(schemaPath string, generatedPath string) error {
	bldr := builder.NewGQLBuilder()

	if err := bldr.ImportPackage(schemaPath); err != nil {
		return fmt.Errorf("generate code: %w", err)
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
