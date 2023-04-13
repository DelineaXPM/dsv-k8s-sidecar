package auth

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
)

const (
	baseAuthWithNSUrl = "%s://dsv-auth.%s/auth"
	hp                = "http"
	hps               = "https"
)

func GetToken(name, ip, brokerNamespace string) (credentials.PerRPCCredentials, error) {
	authUrl := fmt.Sprintf(baseAuthWithNSUrl, hp, brokerNamespace)
	url := util.EnvString("AUTH_URL", authUrl)
	log.WithField("url", url).Info("Fetching Token")
	b := &TokenRequest{
		PodName: name,
		PodIp:   ip,
	}

	body, err := json.Marshal(b)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error marshalling")
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.WithField("error", err.Error()).Error("Error creating request")
		return nil, err
	}

	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json")

	r, err := client.Do(req)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error retrieving data")
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}

	var resp TokenResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		log.WithField("error", err.Error()).Error("Error decoding data ")
		return nil, err
	}

	log.Info("Received Token")
	return NewToken(resp.Token), nil
}

func GetTLsToken(name, ip, brokerNamespace, certFile string) (credentials.PerRPCCredentials, error) {
	authUrl := fmt.Sprintf(baseAuthWithNSUrl, hps, brokerNamespace)
	url := util.EnvString("AUTH_URL", authUrl)
	log.WithField("url", url).Info("Fetching Token")
	b := &TokenRequest{
		PodName: name,
		PodIp:   ip,
	}

	body, err := json.Marshal(b)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error marshaling")
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.WithField("error", err.Error()).Error("Error creating request")
		return nil, err
	}

	dat, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error reading ca file")
		return nil, err
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(dat)
	if !ok {
		panic("failed to parse root certificate")
	}
	tlsConf := &tls.Config{RootCAs: roots}
	tr := &http.Transport{TLSClientConfig: tlsConf}
	client := &http.Client{Transport: tr}

	req.Header.Set("Content-Type", "application/json")

	r, err := client.Do(req)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error retrieving data")
		return nil, err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errors.New(r.Status)
	}

	var resp TokenResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		log.WithField("error", err.Error()).Error("Error decoding data ")
		return nil, err
	}

	log.Info("Received Token")
	return NewToken(resp.Token), nil
}
