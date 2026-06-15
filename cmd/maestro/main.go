package main

import (
	"fmt"
	"os"

	corelogger "github.com/daniildddd/maestro/internal/core/logger"
)

func main() {
	logger, err := corelogger.NewLogger(corelogger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}

	logger.Info("logger successful init")
}
