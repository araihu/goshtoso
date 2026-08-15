// Command scrollregionatchallenge creates the independent random challenge
// which a final T-GS-011 AT capture must echo inside its signed envelope.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const challengeSchema = "goshtoso.t-gs-011.at-challenge.v1"

type captureChallenge struct {
	Schema    string `json:"schema"`
	Challenge string `json:"challenge"`
	IssuedAt  string `json:"issued_at"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func run(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	flags := flag.NewFlagSet("scrollregionatchallenge", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", "", "new challenge JSON path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *output == "" {
		fmt.Fprintln(stderr, "scrollregionatchallenge: --output is required and positional arguments are forbidden")
		return 2
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Fprintf(stderr, "scrollregionatchallenge: read cryptographic randomness: %v\n", err)
		return 1
	}
	challenge := captureChallenge{Schema: challengeSchema, Challenge: hex.EncodeToString(bytes), IssuedAt: now().UTC().Format(time.RFC3339)}
	encoded, err := json.MarshalIndent(challenge, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "scrollregionatchallenge: encode challenge: %v\n", err)
		return 1
	}
	path, err := filepath.Abs(*output)
	if err != nil {
		fmt.Fprintf(stderr, "scrollregionatchallenge: resolve output path: %v\n", err)
		return 1
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		fmt.Fprintf(stderr, "scrollregionatchallenge: create independent challenge output: %v\n", err)
		return 1
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		_ = file.Close()
		fmt.Fprintf(stderr, "scrollregionatchallenge: write challenge: %v\n", err)
		return 1
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		fmt.Fprintf(stderr, "scrollregionatchallenge: sync challenge: %v\n", err)
		return 1
	}
	if err := file.Close(); err != nil {
		fmt.Fprintf(stderr, "scrollregionatchallenge: close challenge: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "T-GS-011 independent AT challenge written to %s\n", path)
	return 0
}
