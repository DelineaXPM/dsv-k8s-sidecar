package secrets

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/cache"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"

	log "github.com/sirupsen/logrus"
)

// List of supported authentication methods.
const (
	clientCreds = "client_credentials"
	cert        = "cert"
	certificate = "certificate"
)

// Paths to a client certificate and private key for authentication by certificate.
const (
	authByCertTLSKey  = "/etc/dsv/certs/tls.key"
	authByCertTLSCert = "/etc/dsv/certs/tls.crt"
)

// Configurations for background goroutine.
const (
	defaultTokenUpdPeriod = 30 * time.Minute
	defaultCacheUpdPeriod = 15 * time.Minute
)

var (
	tr = &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
)

type authBody struct {
	Type               string `json:"grant_type"`
	ID                 string `json:"client_id"`
	Secret             string `json:"client_secret"`
	CertChallengeID    string `json:"cert_challenge_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

type AuthResponse struct {
	Token string `json:"accessToken"`
}

type SecretResponseData struct {
	ID             string                 `json:"id"`
	Path           string                 `json:"path"`
	Type           string                 `json:"type"`
	Attributes     map[string]interface{} `json:"attributes"`
	Value          interface{}            `json:"data"`
	Created        time.Time              `json:"created"`
	LastModified   time.Time              `json:"lastModified"`
	CreatedBy      string                 `json:"createdBy"`
	LastModifiedBy string                 `json:"lastModifiedBy"`
	Version        string                 `json:"version"`
}

type SecretClientError struct {
	Status int
	Error  error
}

type SecretClient interface {
	GetSecret(secret string) (*SecretResponseData, *SecretClientError)
	SetSecretURL(url string)
	Close() error
}

type secretClient struct {
	tenant              string
	id                  string
	secret              string
	token               string
	quit                chan bool
	error               *SecretClientError
	cache               cache.Cache
	baseAuthURL         string
	baseSecretURL       string
	initiateCertAuthURL string
	authType            string
}

func CreateSecretClient(tenant, id, secret, authType string) SecretClient { //nolint:ireturn //ireturn: by design this is ok to keep like this.
	baseURL := util.EnvString("DSV_API_URL", "https://%s.secretsvaultcloud.com/v1")
	baseAuthURL := baseURL + "/token"
	baseSecretURL := baseURL + "/secrets/%s"
	initiateCertAuthURL := baseURL + "/certificate/auth"

	cacheUpdPeriod := defaultCacheUpdPeriod
	if raw := os.Getenv("REFRESH_TIME"); raw != "" {
		var err error
		cacheUpdPeriod, err = time.ParseDuration(raw)
		if err != nil {
			panic("Bad Refresh Time Specified")
		}
	}

	scl := &secretClient{
		tenant:              tenant,
		id:                  id,
		secret:              secret,
		quit:                make(chan bool),
		cache:               cache.CreateMemoryCache(),
		baseAuthURL:         baseAuthURL,
		baseSecretURL:       baseSecretURL,
		initiateCertAuthURL: initiateCertAuthURL,
		authType:            strings.ToLower(authType),
	}

	scl.updateToken()

	go func() {
		defer log.Info("exited timer")

		tokenTicker := time.NewTicker(defaultTokenUpdPeriod)
		defer tokenTicker.Stop()

		cacheTicker := time.NewTicker(cacheUpdPeriod)
		defer cacheTicker.Stop()

		for {
			select {
			case <-tokenTicker.C:
				go scl.updateToken()

			case <-cacheTicker.C:
				go scl.updateCache()

			case <-scl.quit:
				return
			}
		}
	}()

	return scl
}

func (c *secretClient) setError(status int, err error) {
	c.error = &SecretClientError{
		Status: status,
		Error:  err,
	}
}

// TODO Refresh Token
func (c *secretClient) updateToken() {
	var b *authBody
	switch c.authType {
	case clientCreds:
		b = &authBody{
			Type:   clientCreds,
			ID:     c.id,
			Secret: c.secret,
		}

	case cert, certificate:
		challengeID, challenge, err := c.initiateCertAuth()
		if err != nil {
			log.Error("Error creating initiateCertAuth ", err)
			return
		}

		b = &authBody{
			Type:               certificate,
			CertChallengeID:    challengeID,
			DecryptedChallenge: challenge,
		}
	}

	url := fmt.Sprintf(c.baseAuthURL, c.tenant)
	log.WithField("url", url).Info("Fetching Token")

	body, err := json.Marshal(b)
	if err != nil {
		log.Error("Marshal token request body: %v", err)
		c.setError(http.StatusInternalServerError, err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.Error("Create token request: %v", err)
		c.setError(http.StatusInternalServerError, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Token request: %v", err)
		c.setError(http.StatusInternalServerError, err)
		return
	}

	defer resp.Body.Close()

	var respStr AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&respStr); err != nil {
		log.WithField("error", err.Error()).Error("Error decoding data")
		c.setError(http.StatusInternalServerError, err)
	}
	if resp.StatusCode != http.StatusOK {
		c.setError(resp.StatusCode, errors.New(resp.Status))
		return
	}
	c.token = respStr.Token
	log.Info("Received Token")
}

func (c *secretClient) GetSecret(secret string) (*SecretResponseData, *SecretClientError) {
	val := c.cache.Get(secret)
	if val != nil {
		ret, ok := val.(*SecretResponseData)
		if ok {
			return ret, nil
		}
		log.WithField("secret", secret).Infof("Unexpected type in cache: %T", val)
		// Continue as if it is "cache miss" case.
	}

	log.WithField("secret", secret).Info("Cache miss")
	return c.fetchSecretFromDSV(secret)
}

func (c *secretClient) fetchSecretFromDSV(secret string) (*SecretResponseData, *SecretClientError) {
	// If we have an auth error return.
	if c.error != nil {
		return nil, c.error
	}

	url := fmt.Sprintf(c.baseSecretURL, c.tenant, secret)
	log.WithField("url", url).Info("Fetching Secret")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error creating request")
		c.setError(http.StatusInternalServerError, err)
		return nil, c.error
	}

	client := &http.Client{}
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Authorization", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.WithField("error", err.Error()).Error("Error retrieving data")
		c.setError(http.StatusInternalServerError, err)
		return nil, c.error
	}

	defer resp.Body.Close()

	var respStr SecretResponseData
	if err := json.NewDecoder(resp.Body).Decode(&respStr); err != nil {
		log.WithField("error", err.Error()).Error("Error decoding data")
		c.setError(http.StatusInternalServerError, err)
		return nil, c.error
	}
	if resp.StatusCode != http.StatusOK {
		c.setError(resp.StatusCode, errors.New(resp.Status))
		return nil, c.error
	}

	val := &respStr
	c.cache.Set(secret, val)

	return val, nil
}

func (c *secretClient) Close() error {
	c.quit <- true
	return nil
}

func (c *secretClient) updateCache() {
	log.Info("Updating Cache")
	for _, key := range c.cache.KeySet() {
		key := key
		go func() {
			val, err := c.fetchSecretFromDSV(key)
			if err != nil {
				log.WithField("error", err).Error("error updating cache")
				return
			}
			c.cache.Set(key, val)
		}()
	}
}

func (c *secretClient) SetSecretURL(url string) {
	c.baseSecretURL = url
}

func (c *secretClient) initiateCertAuth() (string, string, error) {
	tlsCert, err := os.ReadFile(authByCertTLSCert)
	if err != nil {
		log.WithField("error", err.Error()).Error("Unable to open auth by cert tls cert file")
		return "", "", err
	}

	tlsKey, err := os.ReadFile(authByCertTLSKey)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to open auth by cert tls key file")
		return "", "", err
	}

	block, _ := pem.Decode(tlsKey)
	if block == nil {
		log.Error("unable to decode auth by cert tls  key pem")
		return "", "", errors.New("unable to decode auth by cert key pem")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to parse private key")
		return "", "", err
	}

	request := struct {
		Cert string `json:"client_certificate"`
	}{
		Cert: base64.StdEncoding.EncodeToString(tlsCert),
	}
	response := struct {
		ID        string `json:"cert_challenge_id"`
		Encrypted string `json:"encrypted"`
	}{}

	log.WithField("request", request).Info("initiate auth request")
	url := fmt.Sprintf(c.initiateCertAuthURL, c.tenant)
	log.WithField("url", url).Info("initiateCertAuthUrl")
	serRequest, err := json.Marshal(request)
	if err != nil {
		log.WithField("error", err.Error()).Error("error serializing initiate cert auth request body")
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(serRequest))
	if err != nil {
		log.WithField("error", err.Error()).Error("error creating challenge initiate cert auth request")
		c.setError(http.StatusInternalServerError, err)
		return "", "", err
	}
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.WithField("error", err.Error()).Error("error getting response ")
		return "", "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.WithField("error", err.Error()).Error("error decoding challenge initiate cert response body")
		c.setError(http.StatusInternalServerError, err)
		return "", "", err
	}

	encrypted, err := base64.StdEncoding.DecodeString(response.Encrypted)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to decode challenge initiate cert")
		return "", "", err
	}

	plaintext, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, privateKey, encrypted, nil)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to decrypt challenge initiate cert code")
		return "", "", err
	}

	return response.ID, base64.StdEncoding.EncodeToString(plaintext), nil
}
