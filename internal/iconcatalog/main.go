package iconcatalog

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

// Run executes the icon catalog generator command.
func Run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("iconcatalog", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := Options{}
	fs.StringVar(&opts.CatalogPath, "catalog", "", "schema-v1 catalog path")
	fs.StringVar(&opts.OutputPath, "out", "", "generated Go output path")
	fs.StringVar(&opts.Package, "package", "", "generated Go package")
	fs.StringVar(&opts.Namespace, "namespace", "", "catalog namespace")
	fs.StringVar(&opts.Product, "product", "", "catalog product")
	fs.StringVar(&opts.SpriteURL, "sprite-url", "", "sprite URL metadata")
	fs.StringVar(&opts.ConstPrefix, "const-prefix", "", "generated constant prefix")
	fs.BoolVar(&opts.Check, "check", false, "fail if output is stale")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	if opts.CatalogPath == "" || opts.OutputPath == "" {
		return fmt.Errorf("catalog and out are required")
	}

	f, err := os.Open(opts.CatalogPath)
	if err != nil {
		return fmt.Errorf("open catalog: %w", err)
	}
	catalog, loadErr := Load(f)
	closeErr := f.Close()
	if loadErr != nil {
		return loadErr
	}
	if closeErr != nil {
		return fmt.Errorf("close catalog: %w", closeErr)
	}
	generated, err := Generate(catalog, opts)
	if err != nil {
		return err
	}
	if opts.Check {
		existing, err := os.ReadFile(opts.OutputPath)
		if err != nil || !bytes.Equal(existing, generated) {
			return fmt.Errorf("%s is stale — run `go run ./cmd/iconcatalog` and commit", opts.OutputPath)
		}
		return nil
	}
	if err := os.WriteFile(opts.OutputPath, generated, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "iconcatalog: wrote %s (%d bytes)\n", opts.OutputPath, len(generated))
	return err
}
