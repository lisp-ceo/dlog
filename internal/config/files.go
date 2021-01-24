package config

import (
	"os"
	"path/filepath"

	"github.com/lisp-ceo/dlog/internal/projectpath"
)

var (
	CAFile         = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile  = configFile("server-key.pem")
)

func configFile(filename string) string {
	if dir := os.Getenv("CERT_PATH"); dir != "" {
		return filepath.Join(dir, filename)
	}

	return filepath.Join(projectpath.Root, "_scratch", "certs", filename)
}
