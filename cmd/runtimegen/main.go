// Command runtimegen generates Goshtoso runtime consumers from Muamba's
// acquisition registry and the Goshtoso metadata overlay.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/araihu/assets/assetmeta"
	goshtosoassets "github.com/araihu/goshtoso/assets"
	"github.com/araihu/goshtoso/internal/runtimegen"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "runtimegen: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("runtimegen", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "fail if any generated runtime consumer would change")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	acquisition, err := inventory()
	if err != nil {
		return err
	}
	return runtimegen.Run(root, acquisition, *check, stdout)
}

func inventory() (*assetmeta.Inventory, error) {
	resources := goshtosoassets.MuambaResources()
	adapted := make([]assetmeta.Resource, 0, len(resources))
	for _, resource := range resources {
		item := assetmeta.Resource{Name: resource.Name, Version: resource.Version}
		for _, download := range resource.Downloads {
			item.Downloads = append(item.Downloads, assetmeta.Download{
				Name: download.Name, URL: download.URL, Path: download.Path,
				Integrity: download.Integrity, Hash: download.Hash,
			})
		}
		adapted = append(adapted, item)
	}
	return assetmeta.NewInventory(adapted)
}
