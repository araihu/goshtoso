package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/araihu/goshtoso/internal/jstooling"
)

func main() {
	check := flag.Bool("check", false, "verify generated artifacts without writing")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	results, err := jstooling.Build(*root, *check)
	if err != nil {
		log.Fatal(err)
	}
	mode := "generated"
	if *check {
		mode = "verified"
	}
	for _, result := range results {
		fmt.Printf("%s %s <- %v\n", mode, result.Output, result.Inputs)
	}
}
