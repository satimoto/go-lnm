package util

import (
	"crypto/x509"
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials"
)

func NewCredential(certificate string) (credentials.TransportCredentials, error) {
	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM([]byte(strings.Replace(certificate, "\\n", "\n", -1))) {
		return nil, fmt.Errorf("Error appending certificates")
	}

	return credentials.NewClientTLSFromCert(certPool, ""), nil
}
