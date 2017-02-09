// Package ssl self signs SSL certificates to be used for development use
package ssl

import (
	"crypto/tls"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"math/big"
	"time"
)

// NewSelfSignedTLSConfig returns a new tls config with a self signed certificate
func NewSelfSignedTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour)

	serialNumberMax := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberMax)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore: notBefore,
		NotAfter: notAfter,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{derBytes},
			PrivateKey: key,
			Leaf: template,
		}},
	}, nil
}