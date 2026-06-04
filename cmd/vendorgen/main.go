package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/araihu/goshtoso/internal/vendorgen"
)

func main() {
	if err := vendorgen.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if strings.HasPrefix(err.Error(), "::error::") {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "vendorgen:", err)
		os.Exit(1)
	}
}
