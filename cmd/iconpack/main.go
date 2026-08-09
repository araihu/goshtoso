// Command iconpack generates one consumer-local attributed icon pack from a
// verified Arai Hu Assets release root or archive.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/araihu/goshtoso/iconpack"
)

type stringList []string

func (values *stringList) String() string { return fmt.Sprint([]string(*values)) }

func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "iconpack:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("iconpack", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var opts iconpack.Options
	var names stringList
	fs.StringVar(&opts.ConfigPath, "config", "", "Goshtoso-owned .iconpack.yaml source configuration")
	fs.StringVar(&opts.IconpackLockPath, "lock", "", "Goshtoso-owned .iconpack.lock.yaml path")
	fs.BoolVar(&opts.Trust, "trust", false, "explicitly first-trust unlocked iconpack sources")
	fs.BoolVar(&opts.AllowHTTP, "allow-http", false, "allow HTTP iconpack sources (for local/test endpoints)")
	fs.StringVar(&opts.ReleaseRoot, "release-root", "", "verified Arai Hu Assets extracted release root")
	fs.StringVar(&opts.ReleaseArchive, "release-archive", "", "verified Arai Hu Assets .tar.gz or .zip release archive")
	fs.StringVar(&opts.Release, "release", "", "expected release tag")
	fs.StringVar(&opts.ArchiveSHA256, "archive-sha256", "", "expected release archive SHA-256")
	fs.StringVar(&opts.CatalogSHA256, "catalog-sha256", "", "expected catalog.json SHA-256")
	fs.StringVar(&opts.ReleaseJSONSHA256, "release-json-sha256", "", "expected release.json SHA-256")
	fs.StringVar(&opts.ChecksumsSHA256, "checksums-sha256", "", "expected checksums.txt SHA-256")
	fs.StringVar(&opts.SourceRoot, "source-root", "", "consumer-supplied icon-pack source root")
	fs.StringVar(&opts.SourceArchive, "source-archive", "", "consumer-supplied .tar.gz or .zip icon-pack archive")
	fs.StringVar(&opts.SourceArchiveSHA256, "source-archive-sha256", "", "expected consumer-supplied archive SHA-256")
	fs.StringVar(&opts.SourceManifest, "source-manifest", "", "JSON or YAML icon-pack source manifest")
	fs.Var(&names, "name", "exact catalog canonical name; repeatable")
	fs.StringVar(&opts.SelectionManifest, "manifest", "", "JSON or YAML selection manifest")
	fs.StringVar(&opts.OutputDir, "out", "", "new consumer-owned output directory")
	fs.StringVar(&opts.Package, "package", "", "generated Go package name")
	fs.StringVar(&opts.ConstPrefix, "const-prefix", "Icon", "generated Go constant prefix")
	fs.StringVar(&opts.SpriteURL, "sprite-url", "", "same-origin URL for the generated sprite")
	fs.BoolVar(&opts.Check, "check", false, "verify an existing owned output without publishing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	opts.Names = names
	result, err := iconpack.Generate(ctx, opts)
	if err != nil {
		return err
	}
	verb := "verified"
	if result.Published {
		verb = "published"
	}
	_, err = fmt.Fprintf(os.Stdout, "iconpack: %s %s (%d icons, %s, catalog %s)\n", verb, result.OutputDir, result.SelectedCount, result.Release, result.CatalogSHA256)
	return err
}
