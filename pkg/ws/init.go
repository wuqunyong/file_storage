package ws

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	HttpPort          string
	HttpsPort         string
	ServerCertificate string
	ServerPrivateKey  string
}

func SIGTERMExit() {
	progName := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Warning %s receive process terminal SIGTERM exit 0\n", progName)
}
