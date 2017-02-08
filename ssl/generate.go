package ssl

import (
	"crypto/tls"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"math/big"
	"crypto/x509/pkix"
	"fmt"
	"time"
)

func NewSelfSignedTLSConfig(org string, duration time.Duration) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("/secrets/ssl/tls.crt", "/secrets/ssl/tls.key")
	if err != nil {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %v", err)
		}

		notBefore := time.Now()
		notAfter := notBefore.Add(duration)

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to generate serial number: %s", err)
		}

		template := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{org},
			},
			NotBefore: notBefore,
			NotAfter: notAfter,

			KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
		if err != nil {
			return nil, fmt.Errorf("failed to create certificate: %s", err)
		}

		cert.Certificate = [][]byte{derBytes}
		cert.PrivateKey = key
		cert.Leaf = template
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}, nil
}