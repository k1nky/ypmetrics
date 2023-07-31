package main

import (
	"time"

	"github.com/k1nky/ypmetrics/internal/agent"
)

func main() {
	parseFlags()
	a := agent.New(agent.WithEndpoint(address.String()), agent.WithPollInterval(time.Duration(pollInterval)), agent.WithReportInterval(time.Duration(reportInterval)))
	a.Run()
}
