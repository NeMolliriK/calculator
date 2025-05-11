package main

import (
	"calculator/internal/agent"
	"calculator/pkg/loggers"
)

func main() {
	loggers.InitLogger("agent", "agent_logs.txt")
	agent.Run()
}
