package main

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

func fetchCerts(endpoints ...string) (string, error) {
	out := bytes.Buffer{}

	for _, endpoint := range endpoints {
		parts := strings.Split(endpoint, ":")
		if len(parts) == 1 {
			endpoint = endpoint + ":443"
		}
		log.Printf("Fetching certs for %q\n", endpoint)
		conn, err := tls.Dial("tcp", endpoint, &tls.Config{})
		if err != nil {
			return "", fmt.Errorf("failed to connect: " + err.Error())
		}
		if err := conn.Close(); err != nil {
			return "", err
		}

		for _, crt := range conn.ConnectionState().PeerCertificates {
			pem.Encode(&out, &pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})
		}
	}

	return string(out.Bytes()), nil
}

var t = template.Must(template.New("pkg").Parse("package {{.Package}}\n// {{.Variable}} contains certificates for {{.Domains}}\nvar {{.Variable}} = []byte(`{{.Certificates}}`)"))

type data struct {
	Package      string
	Certificates string
	Domains      string
	Variable     string
}

func main() {
	pkg := flag.String("pkg", "", "package to generate certs into")
	file := flag.String("o", "", "file to generate certs into")
	exported := flag.Bool("exported", false, "export the certificate variable")
	flag.Parse()

	if pkg == nil || file == nil || *pkg == "" || *file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	certs, err := fetchCerts(flag.Args()...)
	if err != nil {
		panic(err)
	}

	var bs = bytes.Buffer{}
	variable := "certs"
	if *exported {
		variable = "Certs"
	}
	if err := t.Execute(&bs, data{
		Package:      *pkg,
		Certificates: certs,
		Domains:      strings.Join(flag.Args(), ", "),
		Variable:     variable,
	}); err != nil {
		panic(err)
	}

	b, err := format.Source(bs.Bytes())
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(*file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		panic(err)
	}
}
