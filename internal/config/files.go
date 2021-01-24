package config

import (
	"os"
	"path/filepath"

	"github.com/lisp-ceo/dlog/internal/projectpath"
)

var (
	CAFile         = certsFile("ca.pem")

	ServerCertFile = certsFile("server.pem")
	ServerKeyFile  = certsFile("server-key.pem")

	RootClientCertFile = certsFile("root.pem")
	RootClientKeyFile = certsFile("root-key.pem")
	UnauthorizedCertFile = certsFile("unauthorized.pem")
	UnauthorizedKeyFile = certsFile("unauthorized-key.pem")

	ACLModelFile = authFile("model.conf")
	ACLPolicyFile = authFile("policy.csv")
)

func certsFile(filename string) string {
	if dir := os.Getenv("CERT_PATH"); dir != "" {
		return filepath.Join(dir, filename)
	}

	return filepath.Join(projectpath.Root, "_scratch", "certs", filename)
}

func authFile(filename string) string {
	if dir := os.Getenv("CERT_PATH"); dir != "" {
		return filepath.Join(dir, filename)
	}

	return filepath.Join(projectpath.Root, "_scratch", "auth", filename)
}
