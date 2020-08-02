package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{Use: "gentc"}
	cmd.AddCommand(
		&cobra.Command{
			Use:     "init",
			Short:   "initialize a GraphQL project schema",
			Example: "gentc init",
			Run: func(cmd *cobra.Command, _ []string) {
				if err := initProj(defaultSchemaPath); err != nil {
					log.Fatalln(err)
				}
			},
		},
	)

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func initProj(target string) error {
	if err := os.MkdirAll(target, os.ModePerm); err != nil {
		return err
	}

	if err := writeFileIfNotExist("graph/schema/query.go", []byte(defaultQuery), 0644); err != nil {
		return fmt.Errorf("creating query.go file: %w", err)
	}

	if err := writeFileIfNotExist("graph/generate.go", []byte(genFile), 0644); err != nil {
		return fmt.Errorf("creating generate.go file: %w", err)
	}

	return nil
}

func writeFileIfNotExist(filename string, data []byte, perm os.FileMode) error {
	_, err := os.Stat(filename)

	if err == nil || !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists", filename)
	}

	return ioutil.WriteFile(filename, data, perm)
}

const (
	defaultSchemaPath = "./graph/schema"
	genFile           = `package graph

//go:generate go run github.com/sijad/gentle/cmd/gentc generate ./schema
`
	defaultQuery = `package schema

// Query holds the GraphQL Query definition.
type Query struct{}

// Hello says hello to world
func (Query) Hello() string {
	return "world"
}
`
)
