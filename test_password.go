//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	fmt.Print("Password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Length: %d\n", len(pw))
	fmt.Printf("Bytes:  %v\n", pw)
	fmt.Printf("String: %q\n", string(pw))
}
