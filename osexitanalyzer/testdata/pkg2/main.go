package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	f()
}

func f() {
	os.Exit(1)
}
