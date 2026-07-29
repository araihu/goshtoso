// Command iconcatalog generates typed sprite bindings from a schema-v1 asset
// catalog.
package main

import (
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/iconcatalog"
)

func main() {
	if err := iconcatalog.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "iconcatalog:", err)
		os.Exit(1)
	}
}
