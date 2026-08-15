package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/araihu/goshtoso/site/internal/e2econstraints"
)

func main() {
	compileMatrix := flag.Bool("compile-matrix", false, "compile every focused E2E identity")
	listMatrix := flag.Bool("list-matrix", false, "compare go test -list with the parsed identity inventory")
	printManifest := flag.Bool("print-manifest", false, "print the inventory derived from current constraints")
	printSpecializedTests := flag.String("print-specialized-tests", "", "print one specialized suite's manifest-owned selected tests")
	flag.Parse()

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

	bareCommands, err := e2econstraints.FindBareE2ECommands("..")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, finding := range bareCommands {
		fmt.Fprintf(os.Stderr, "%s:%d: bare E2E command: %s\n", finding.Path, finding.Line, finding.Text)
	}
	if len(bareCommands) > 0 {
		os.Exit(1)
	}

	manifest, err := e2econstraints.LoadManifest("tests/e2e/identities.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	suite, err := e2econstraints.InspectSuite("tests/e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := e2econstraints.ValidateSuite(suite, manifest); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *printSpecializedTests != "" {
		selected, err := e2econstraints.SpecializedSelectedTests(manifest, *printSpecializedTests)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, test := range selected {
			fmt.Fprintln(os.Stdout, test)
		}
		return
	}
	if *printManifest {
		if err := e2econstraints.WriteManifest(os.Stdout, e2econstraints.ExpandedManifest(suite, manifest)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *compileMatrix {
		if err := e2econstraints.RunCompileMatrix(context.Background(), ".", manifest, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *listMatrix {
		if err := e2econstraints.RunListMatrix(context.Background(), ".", suite, manifest, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
