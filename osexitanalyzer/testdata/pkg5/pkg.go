package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	os.Exit(1) // want `avoid calling os.Exit in the main function`
}
