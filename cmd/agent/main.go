package main

import (
	"github.com/k1nky/ypmetrics/internal/agent"
)

var (
	config *Config
)

func main() {
	var err error

	config, err = Parse(nil)
	if err != nil {
		panic(err)
	}
	a := agent.New(agent.WithEndpoint(config.Address.String()), agent.WithPollInterval(config.PollInterval), agent.WithReportInterval(config.ReportInterval))
	a.Run()
}
