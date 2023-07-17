package main

import (
	"fmt"
	"os"
)

func mainImpl() error {
	fmt.Println("let's make it clean bröther")
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
