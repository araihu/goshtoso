package main

import (
	"encoding/json"
	"fmt"

	"github.com/araihu/goshtoso/assets"
)

func main() {
	identity, err := json.Marshal(assets.GoshtosoVersion())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(identity))
}
