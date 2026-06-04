package main

import (
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/themegen"
)

func main() {
	if err := themegen.Run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "themegen:", err)
		os.Exit(1)
	}
}
