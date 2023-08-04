package main

import (
	"fmt"
	"os"

	"github.com/k1nky/ypmetrics/internal/agent"
	"github.com/k1nky/ypmetrics/internal/config"
)

func main() {
	cfg := config.AgentConfig{}
	if err := config.ParseAgentConfig(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	a := agent.New(cfg)
	a.Run()
}
