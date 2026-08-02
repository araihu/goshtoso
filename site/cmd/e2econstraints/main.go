package main

import (
	"fmt"
	"os"

	"github.com/araihu/goshtoso/site/internal/e2econstraints"
)

func main() {
	findings, err := e2econstraints.FindCrossFileDeclarations(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, finding := range findings {
		fmt.Fprintf(os.Stderr, "%s: %s used by %v\n", finding.DeclaringFile, finding.Name, finding.ConsumerFiles)
	}
	if len(findings) > 0 {
		os.Exit(1)
	}
}
