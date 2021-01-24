package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		// verify the client's certificate
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, err
		}
	}

	if cfg.CAFile != "" {
		b, err := ioutil.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}

		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM(b)

		if !ok {
			return nil, ErrFailedParsingRootCA
		}

		if cfg.Server {
			// which CA to use to verify the client certificate
			tlsConfig.ClientCAs = ca

			// client certificate is required to be valid
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			// set client's root CAs
			tlsConfig.RootCAs = ca
		}

		tlsConfig.ServerName = cfg.ServerAddress
	}

	return tlsConfig, nil
}

var (
	ErrFailedParsingRootCA = errors.New("failed to parse root certificate")
)
