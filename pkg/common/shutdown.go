package common

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/wuqunyong/file_storage/pkg/logger"
)

func WaitForShutdown() {
	// Wait for the process to be shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	SIGTERMExit()
}

func SIGTERMExit() {
	progName := filepath.Base(os.Args[0])
	logger.Log(logger.WarnLevel, "SIGTERMExit", "value", fmt.Sprintf("Warning %s receive process terminal SIGTERM exit 0", progName))
}
