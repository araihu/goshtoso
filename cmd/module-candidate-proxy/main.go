// Command module-candidate-proxy builds an authenticated offline Go module proxy.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/modulecandidate"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, modulecandidate.Build))
}

type buildCandidate func(context.Context, modulecandidate.Config) (modulecandidate.Result, error)

func run(args []string, stdout, stderr *os.File, build buildCandidate) int {
	var config modulecandidate.Config
	flags := flag.NewFlagSet("module-candidate-proxy", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&config.Repository, "repository", "", "Git repository containing the exact candidate")
	flags.StringVar(&config.ModulePath, "module-path", "", "expected Go module path")
	flags.StringVar(&config.Commit, "commit", "", "exact 40-character candidate commit")
	flags.StringVar(&config.ExpectedTree, "tree", "", "exact 40-character candidate tree")
	flags.StringVar(&config.Subdir, "subdir", "", "module subdirectory; only empty is supported")
	flags.StringVar(&config.Output, "output", "", "empty output directory for the file proxy")
	flags.StringVar(&config.DependencyProxy, "dependency-proxy", "", "explicit file:// proxy containing dependency artifacts")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "module-candidate-proxy: positional arguments are forbidden")
		return 2
	}
	seen := map[string]bool{}
	flags.Visit(func(value *flag.Flag) { seen[value.Name] = true })
	for _, required := range []string{"repository", "module-path", "commit", "tree", "subdir", "output", "dependency-proxy"} {
		if !seen[required] {
			fmt.Fprintf(stderr, "module-candidate-proxy: -%s is required\n", required)
			return 2
		}
	}
	if config.Repository == "" || config.ModulePath == "" || config.Commit == "" || config.ExpectedTree == "" || config.Output == "" || config.DependencyProxy == "" {
		fmt.Fprintln(stderr, "module-candidate-proxy: required flag values must not be empty except -subdir")
		return 2
	}
	result, err := build(context.Background(), config)
	if err != nil {
		fmt.Fprintln(stderr, "module-candidate-proxy:", err)
		return 1
	}
	var encoded bytes.Buffer
	if err := json.NewEncoder(&encoded).Encode(result); err != nil {
		fmt.Fprintln(stderr, "module-candidate-proxy: encode result:", err)
		return 1
	}
	if _, err := stdout.Write(encoded.Bytes()); err != nil {
		fmt.Fprintln(stderr, "module-candidate-proxy: write result:", err)
		return 1
	}
	return 0
}
