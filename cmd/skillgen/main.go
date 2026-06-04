package main

import (
	"fmt"
	"os"

	"github.com/araihu/goshtoso/internal/skillgen"
)

func main() {
	if err := skillgen.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "skillgen:", err)
		os.Exit(1)
	}
}
