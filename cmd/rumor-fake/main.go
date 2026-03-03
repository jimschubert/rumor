package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/jimschubert/rumor/internal/faker"
	"github.com/jimschubert/rumor/internal/store/jsonstore"
)

//go:embed schemas/*.json
var embeddedSchemas embed.FS

var (
	programName = "rumor-fake"
	version     = "dev"
	commit      = "unknown SHA"
)

var CLI struct {
	Schema           string           `arg:"" help:"Path to JSON schema file defining the data structure"`
	Count            int              `short:"c" default:"10" help:"Number of records to generate"`
	Resource         string           `short:"r" help:"Resource name (defaults to schema filename without extension)"`
	TrimSchemaSuffix bool             `short:"t" default:"false" help:"Trim .schema suffix from resource name if present (e.g. users.schema.json -> users)"`
	Output           string           `short:"o" default:"db.json" help:"Output JSON database file"`
	Version          kong.VersionFlag `short:"v" help:"Print version information"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name(programName),
		kong.Description("Generate fake data for the rumor server based on a schema definition."),
		kong.UsageOnError(),
		kong.Vars{"version": fmt.Sprintf("%s (%s)", version, commit)},
	)

	if err := run(); err != nil {
		ctx.Errorf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	schemaReader, err := createSchemaReader(CLI.Schema)
	if err != nil {
		return fmt.Errorf("failed to create schema reader: %w", err)
	}

	defer func() {
		_ = schemaReader.Close()
	}()

	schema, err := faker.LoadSchema(schemaReader)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	resource := CLI.Resource
	if resource == "" {
		resource = filepath.Base(CLI.Schema)
		resource, _ = strings.CutSuffix(resource, ".json")
		if CLI.TrimSchemaSuffix {
			resource, _ = strings.CutSuffix(resource, ".schema")
		}
	}

	store, err := jsonstore.New(CLI.Output)
	if err != nil {
		return fmt.Errorf("failed to load store: %w", err)
	}

	for i := 0; i < CLI.Count; i++ {
		record := faker.GenerateRecord(schema)
		_, err := store.Create(resource, record)
		if err != nil {
			return fmt.Errorf("failed to create record %d: %w", i+1, err)
		}
	}

	fmt.Printf("Generated %d fake records in resource %q to %s\n", CLI.Count, resource, CLI.Output)
	return nil
}

func createSchemaReader(schemaPath string) (io.ReadCloser, error) {
	schemaFile, err := os.Open(schemaPath)
	if err == nil {
		return schemaFile, nil
	}

	// fallback to embedded for the schema by name (only)
	nameOnly := filepath.Base(schemaPath)
	embeddedPath := filepath.Join("schemas", nameOnly)
	embeddedFile, embedErr := embeddedSchemas.Open(embeddedPath)
	if embedErr != nil {
		return nil, fmt.Errorf("failed to open schema file locally or from embedded: %w", err)
	}
	log.Printf("fyi: schema file not found at %s, using embedded version of the same name: %s", schemaPath, nameOnly)
	return embeddedFile, nil
}
