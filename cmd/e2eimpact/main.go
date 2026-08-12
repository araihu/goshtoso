package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/e2eimpact"
)

func main() {
	base := flag.String("base", "", "ancestor commit at the start of the change range")
	head := flag.String("head", "HEAD", "commit at the end of the change range")
	changesFile := flag.String("changes-file", "", "NUL-delimited git diff --name-status -z -M stream")
	flag.Parse()
	var result e2eimpact.Result
	if *changesFile != "" {
		data, err := os.ReadFile(*changesFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		result = e2eimpact.SelectNameStatus(context.Background(), ".", data)
	} else {
		result = e2eimpact.Select(context.Background(), ".", *base, *head)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
