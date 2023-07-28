package main

import "github.com/k1nky/ypmetrics/internal/agent"

func main() {
	a := agent.New()
	a.Run()
}
