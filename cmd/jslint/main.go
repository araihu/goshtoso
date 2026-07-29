package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/araihu/goshtoso/internal/jstooling"
)

func main() {
	baselinePath := flag.String("inline-baseline", "tools/javascript/inline-baseline.txt", "inline JavaScript baseline path")
	inventory := flag.Bool("inventory", false, "print every inline JavaScript finding")
	root := flag.String("root", ".", "repository root")
	threshold := flag.Float64("similarity-threshold", 0.82, "minimum structural similarity score")
	updateBaseline := flag.Bool("update-inline-baseline", false, "replace baseline with current finding fingerprints")
	flag.Parse()

	sources, err := jstooling.SourceFiles(*root)
	failIf(err)
	failIf(jstooling.ValidateJavaScript(sources))

	pairs := jstooling.SimilarFunctions(sources, *threshold)
	for _, pair := range pairs {
		fmt.Printf("similar %.3f %s:%d:%s %s:%d:%s\n",
			pair.Score,
			pair.Left.Path, pair.Left.Line, pair.Left.Name,
			pair.Right.Path, pair.Right.Line, pair.Right.Name,
		)
	}
	if len(pairs) == 0 {
		fmt.Printf("similarity: no pairs at or above %.3f\n", *threshold)
	}

	findings, err := jstooling.ScanInlineJavaScript(*root)
	failIf(err)
	if *updateBaseline {
		keys := make([]string, 0, len(findings))
		for _, finding := range findings {
			keys = append(keys, finding.Key())
		}
		sort.Strings(keys)
		var content strings.Builder
		content.WriteString("# Existing extraction candidates. Regenerate only after reviewing inventory.\n")
		for _, key := range keys {
			content.WriteString(key + "\n")
		}
		path := filepath.Join(*root, filepath.FromSlash(*baselinePath))
		failIf(os.MkdirAll(filepath.Dir(path), 0o755))
		failIf(os.WriteFile(path, []byte(content.String()), 0o644))
		fmt.Printf("inline baseline: wrote %d findings to %s\n", len(findings), *baselinePath)
		return
	}

	baseline, err := jstooling.ReadBaseline(filepath.Join(*root, filepath.FromSlash(*baselinePath)))
	failIf(err)
	if *inventory {
		for _, finding := range findings {
			fmt.Printf("inline %s:%d [%s] %s\n", finding.Path, finding.Line, finding.Kind, finding.Summary)
		}
	}
	newFindings := jstooling.NewFindings(findings, baseline)
	for _, finding := range newFindings {
		fmt.Printf("new inline %s:%d [%s] %s\n", finding.Path, finding.Line, finding.Kind, finding.Summary)
	}
	if len(newFindings) > 0 {
		log.Fatalf("inline JavaScript policy: %d new extraction candidate(s)", len(newFindings))
	}
	fmt.Printf("inline JavaScript: %d baselined extraction candidate(s), no new findings\n", len(findings))
}

func failIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
