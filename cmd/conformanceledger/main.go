package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/araihu/goshtoso/internal/conformanceledger"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("conformanceledger", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repo := flags.String("repo", "", "exact Goshtoso repository root")
	stateFixturesOutput := flags.String("state-fixtures-output", "", "generate exhaustive E2E state fixture Go source")
	siteLedgerOutput := flags.String("site-ledger-output", "", "generate site-module-local E2E ledger Go source directory")
	commit := flags.String("commit", "", "exact source commit")
	tree := flags.String("tree", "", "exact source tree")
	receiptsPath := flags.String("receipts", "", "JSON array of authenticated prior receipts")
	executionsPath := flags.String("executions", "", "optional authenticated execution receipt envelope JSON")
	executionsSHA256 := flags.String("executions-sha256", "", "required SHA-256 of authenticated execution receipt envelope")
	atBlocker := flags.String("at-blocker", "", "real AT capability blocker receipt")
	output := flags.String("output", "", "generated ledger JSON path")
	requireClosure := flags.Bool("require-closure", false, "fail unless every execution row is executed or justified not applicable")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *stateFixturesOutput != "" {
		if *repo == "" {
			return fmt.Errorf("-repo is required with -state-fixtures-output")
		}
		source, err := conformanceledger.GenerateStateFixtureSource(*repo)
		if err != nil {
			return err
		}
		if err := os.WriteFile(*stateFixturesOutput, source, 0o600); err != nil {
			return fmt.Errorf("write state fixtures: %w", err)
		}
		fmt.Fprintf(stdout, "state_fixtures=%s\n", *stateFixturesOutput)
		return nil
	}
	if *siteLedgerOutput != "" {
		if *repo == "" {
			return fmt.Errorf("-repo is required with -site-ledger-output")
		}
		if err := conformanceledger.WriteSiteLedgerSource(*repo, *siteLedgerOutput); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "site_ledger=%s\n", *siteLedgerOutput)
		return nil
	}
	if *repo == "" || *commit == "" || *tree == "" || *receiptsPath == "" || *atBlocker == "" || *output == "" {
		return fmt.Errorf("-repo, -commit, -tree, -receipts, -at-blocker, and -output are required")
	}

	receiptBytes, err := os.ReadFile(*receiptsPath)
	if err != nil {
		return fmt.Errorf("read receipts: %w", err)
	}
	var receipts []conformanceledger.ReceiptInput
	if err := json.Unmarshal(receiptBytes, &receipts); err != nil {
		return fmt.Errorf("parse receipts: %w", err)
	}
	ledger, inventory, err := conformanceledger.GenerateSkeleton(conformanceledger.GenerationConfig{
		RepoRoot:         *repo,
		SourceCommit:     *commit,
		SourceTree:       *tree,
		Receipts:         receipts,
		ATBlockerReceipt: *atBlocker,
	})
	if err != nil {
		return err
	}
	if *executionsPath != "" || *executionsSHA256 != "" {
		if *executionsPath == "" || *executionsSHA256 == "" {
			return fmt.Errorf("-executions and -executions-sha256 must be provided together")
		}
		if err := conformanceledger.ReadAndApplyExecutionReceiptEnvelope(&ledger, *executionsPath, *executionsSHA256); err != nil {
			return err
		}
		if err := conformanceledger.Validate(ledger, inventory); err != nil {
			return fmt.Errorf("validate ledger after execution receipts: %w", err)
		}
	}
	if *requireClosure {
		if err := conformanceledger.ValidateClosure(ledger, inventory); err != nil {
			return err
		}
	}
	encoded, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("encode ledger: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o600); err != nil {
		return fmt.Errorf("write ledger: %w", err)
	}
	fmt.Fprintf(stdout, "ledger=%s rows=%d source_commit=%s source_tree=%s\n", *output, len(ledger.Rows), ledger.SourceCommit, ledger.SourceTree)
	return nil
}
