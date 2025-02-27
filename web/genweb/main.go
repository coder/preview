package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/xerrors"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
)

func main() {
	outFile := flag.String("out", "", "output file")
	flag.Parse()

	gen, err := guts.NewGolangParser()
	if err != nil {
		log.Fatalf("new convert: %v", err)
	}

	generateDirectories := map[string]string{
		"github.com/coder/preview/web":   "",
		"github.com/coder/preview/types": "",
	}
	for dir, prefix := range generateDirectories {
		err = gen.IncludeGenerateWithPrefix(dir, prefix)
		if err != nil {
			log.Fatalf("include generate package %q: %v", dir, err)
		}
	}

	referencePackages := map[string]string{}
	for pkg, prefix := range referencePackages {
		err = gen.IncludeReference(pkg, prefix)
		if err != nil {
			log.Fatalf("include reference package %q: %v", pkg, err)
		}
	}

	err = TypeMappings(gen)
	if err != nil {
		log.Fatalf("type mappings: %v", err)
	}

	ts, err := gen.ToTypescript()
	if err != nil {
		log.Fatalf("to typescript: %v", err)
	}

	TsMutations(ts)

	output, err := ts.Serialize()
	if err != nil {
		log.Fatalf("serialize: %v", err)
	}

	if outFile != nil {
		_ = os.WriteFile(*outFile, []byte(output), 0644)
	} else {
		_, _ = fmt.Println(output)
	}
}

func TsMutations(ts *guts.Typescript) {
	ts.ApplyMutations(
		// Enum list generator
		config.EnumLists,
		// Export all top level types
		config.ExportTypes,
		// Readonly interface fields
		config.ReadOnly,
		// Add ignore linter comments
		config.BiomeLintIgnoreAnyTypeParameters,
		// Omitempty + null is just '?' in golang json marshal
		// number?: number | null --> number?: number
		config.SimplifyOmitEmpty,
		config.NullUnionSlices,
	)
}

// TypeMappings is all the custom types for codersdk
func TypeMappings(gen *guts.GoParser) error {
	gen.IncludeCustomDeclaration(config.StandardMappings())

	gen.IncludeCustomDeclaration(map[string]guts.TypeOverride{})

	err := gen.IncludeCustom(map[string]string{
		"github.com/coder/preview/types.HCLString": "string",
		// Serpent fields should be converted to their primitive types
		"github.com/coder/serpent.Regexp":         "string",
		"github.com/coder/serpent.StringArray":    "string",
		"github.com/coder/serpent.String":         "string",
		"github.com/coder/serpent.YAMLConfigPath": "string",
		"github.com/coder/serpent.Strings":        "[]string",
		"github.com/coder/serpent.Int64":          "int64",
		"github.com/coder/serpent.Bool":           "bool",
		"github.com/coder/serpent.Duration":       "int64",
		"github.com/coder/serpent.URL":            "string",
		"github.com/coder/serpent.HostPort":       "string",
		"encoding/json.RawMessage":                "map[string]string",
	})
	if err != nil {
		return xerrors.Errorf("include custom: %w", err)
	}

	return nil
}
