package main

import (
	"fmt"
	"os"

	core_logger "github.com/daniildddd/maestro/internal/core/logger"
)

func main() {
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}

	logger.Info("logger successful init")
}
