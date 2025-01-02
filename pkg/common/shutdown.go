package common

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
	slog.Warn("SIGTERMExit", "value", fmt.Sprintf("Warning %s receive process terminal SIGTERM exit 0", progName))
}
