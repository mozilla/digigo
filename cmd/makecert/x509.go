package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func makeCSRAndKey() (csr *x509.CertificateRequest, csrPEM, privKeyPEM string, err error) {
	var priv interface{}
	switch *ecdsaCurve {
	case "":
		priv, err = rsa.GenerateKey(rand.Reader, *rsaBits)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized elliptic curve: %q", *ecdsaCurve)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}
	hosts := strings.Split(*host, ",")
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         hosts[0],
			Organization:       []string{*org},
			OrganizationalUnit: []string{*ou},
			Country:            []string{*country},
			Province:           []string{*st},
			Locality:           []string{*loc},
		},
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	csrB, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		err = errors.Wrap(err, "failed to generate CSR")
		return
	}
	csr, err = x509.ParseCertificateRequest(csrB)
	if err != nil {
		err = errors.Wrap(err, "failed to parse CSR")
		return
	}
	// convert the CSR bytes to PEM
	var csrBuf bytes.Buffer
	pem.Encode(&csrBuf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrB})
	csrPEM = string(csrBuf.Bytes())
	// convert the private key to PEM
	privKeyPEM, err = getPrivateKeyPEM(priv)
	return csr, csrPEM, privKeyPEM, err
}

func getPrivateKeyPEM(priv interface{}) (string, error) {
	var buf bytes.Buffer
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		pem.Encode(&buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)})
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return "", errors.Wrap(err, "unable to marshal ECDSA private key")
		}
		pem.Encode(&buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	default:
		return "", errors.Errorf("unknown private key type %T", priv)
	}
	return string(buf.Bytes()), nil
}
