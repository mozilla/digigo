package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"os"

	"go.mozilla.org/digigo"
)

var (
	debug      = flag.Bool("D", false, "Enable debug output")
	host       = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	ou         = flag.String("ou", "Cloud Services", "Organizational Unit")
	org        = flag.String("org", "Mozilla Corporation", "Organization")
	loc        = flag.String("loc", "Mountain View", "Locality")
	st         = flag.String("st", "California", "State")
	country    = flag.String("c", "US", "Country")
	email      = flag.String("email", "hostmaster@mozilla.com", "Email")
	rsaBits    = flag.Int("rsa-bits", 2048, "Size of RSA key to generate, default to rsa 2048, ignored if --ecdsa-curve is set")
	ecdsaCurve = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key, valid values are P256 and P384")
	valYears   = flag.Int("validity-years", 1, "Years of validity of the signed certificate: 1 (default), 2 or 3")
)

func main() {
	flag.Parse()
	if len(*host) == 0 {
		log.Fatalf("Missing required --host parameter")
	}
	// Create and test the connection to Digicert
	cli, err := digigo.NewClient(os.Getenv("DIGICERT_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	if *debug {
		cli.EnableDebug()
	}
	_, err = cli.ViewProductList()
	if err != nil {
		log.Fatal(err)
	}

	// Step 1: make a certificate request and a private key
	csr, csrPEM, keyPEM, err := makeCSRAndKey()
	if err != nil {
		log.Fatal(err)
	}
	if *debug {
		fmt.Printf("%s\n%s\n", csrPEM, keyPEM)
	}

	// Step 2: submit an order to digicert
	var order digigo.Order
	order.Certificate.CommonName = csr.Subject.CommonName
	order.Certificate.DNSNames = csr.DNSNames
	order.Certificate.Csr = csrPEM
	order.Certificate.OrganizationUnits = csr.Subject.OrganizationalUnit
	order.Certificate.ServerPlatform.ID = 45 // nginx, aka. standard PEM
	order.ValidityYears = *valYears
	switch csr.SignatureAlgorithm {
	case x509.SHA384WithRSA, x509.ECDSAWithSHA384:
		order.Certificate.SignatureHash = "sha384"
	case x509.SHA512WithRSA, x509.ECDSAWithSHA512:
		order.Certificate.SignatureHash = "sha512"
	default:
		order.Certificate.SignatureHash = "sha256"
	}
	// FIXME: actually call ListOrganizations
	order.Organization.ID = 147486 // mozilla org ID
	orderID, err := cli.SubmitOrder(order, "ssl")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("placed order with ID", orderID)
}
