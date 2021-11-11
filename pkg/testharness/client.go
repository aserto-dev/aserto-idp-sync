package testharness

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
)

// CreateClient creates a new http client that can talk to the API
func (h *TestHarness) CreateClient() *http.Client {
	caCert, err := ioutil.ReadFile(h.Idpsync.Configuration.API.Gateway.Certs.TLSCACertPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS12,
			}}}

	return client
}
