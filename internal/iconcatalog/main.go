package iconcatalog

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var atomicRename = os.Rename

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
	if err := writeAtomic(opts.OutputPath, generated); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "iconcatalog: wrote %s (%d bytes)\n", opts.OutputPath, len(generated))
	return err
}

func writeAtomic(path string, contents []byte) (err error) {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temporary.Close()
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o644); err != nil {
		return fmt.Errorf("set temporary output permissions: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		return fmt.Errorf("write temporary output: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary output: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := atomicRename(temporaryPath, path); err != nil {
		return fmt.Errorf("rename output: %w", err)
	}
	committed = true
	return nil
}
