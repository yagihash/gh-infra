package main

import (
	"fmt"
	"os"
)

var (
	version  = "dev"
	revision = "HEAD"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("gh-infra %s (%s)\n", version, revision)
		return
	}

	fmt.Println("gh-infra")
}
