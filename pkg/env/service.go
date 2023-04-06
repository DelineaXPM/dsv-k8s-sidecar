package env

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/secrets"
	"github.com/DelineaXPM/dsv-k8s-sidecar/pkg/util"

	log "github.com/sirupsen/logrus"
)

//nolint:gochecknoglobals // already in design, leaving as ok
var (
	configDir         = util.EnvString("CONFIG_DIR", "/tmp/secret/")
	SecretFile        = configDir + util.EnvString("SECRET_FILE", "dsv.json")
	refreshTimeString = util.EnvString("REFRESH_TIME", "15m")
)

const SecretEnvName = "DSV_SECRETS"

type EnvironmentAgent interface {
	Run() <-chan error
	UpdateEnv()
	Close()
}

type environmentAgent struct {
	vars   []string
	stop   chan bool
	client secrets.DsvClient
}

func CreateEnvironmentAgent(client secrets.DsvClient) EnvironmentAgent {
	envString := os.Getenv(SecretEnvName)

	// cleanup trailing spaces
	var vars []string
	for _, v := range strings.Split(envString, " ") {
		vars = append(vars, strings.TrimSpace(v))
	}

	return &environmentAgent{
		vars:   vars,
		client: client,
	}
}

func (a *environmentAgent) Run() <-chan error {
	refreshTime, err := time.ParseDuration(refreshTimeString)
	if err != nil {
		log.Error("Invalid refresh time, defaulting to 15m")
		refreshTime, _ = time.ParseDuration("15m")
	}
	log.Info("Running.....")
	ticker := time.NewTicker(refreshTime)
	errs := make(chan error)

	a.UpdateEnv()
	go func() {
	UpdateLoop:
		for {
			select {
			case <-ticker.C:
				go a.UpdateEnv()
			case <-a.stop:
				break UpdateLoop
			}
		}
	}()

	return errs
}

func (a *environmentAgent) UpdateEnv() {
	a.write(a.fetch(a.vars))
}

func (a *environmentAgent) fetch(keys []string) chan struct {
	key   string
	value interface{}
} {
	var wg sync.WaitGroup
	wg.Add(len(a.vars))

	out := make(chan struct {
		key   string
		value interface{}
	})
	for _, key := range keys {
		go func(k string) {
			log.WithField("key", k).Info("Fetching env variable")
			resp, err := a.client.GetSecret(context.Background(), &secrets.Secret{Path: k})
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
					"path":  k,
				}).Fatal("could not fetch secret")
			}

			if resp.Type == "json" {
				var v map[string]interface{}
				err := json.Unmarshal([]byte(resp.Value), &v)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Fatal("unable to unmarshal")
				}
				out <- struct {
					key   string
					value interface{}
				}{k, v}
			} else {
				out <- struct {
					key   string
					value interface{}
				}{k, resp.Value}
			}

			wg.Done()
		}(key)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (a *environmentAgent) write(secrets chan struct {
	key   string
	value interface{}
},
) error {
	m := make(map[string]interface{})

	for secret := range secrets {
		m[secret.key] = secret.value
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("unable to marshal map")
		return err
	}

	if _, err := os.Stat(configDir); err != nil {
		err := os.MkdirAll(configDir, os.ModePerm)
		if err != nil {
			log.WithField("error", err.Error()).Fatal("unable to create secrets directory")
			return err
		}
	}

	if err := os.WriteFile(SecretFile, data, os.ModePerm); err != nil {
		log.WithField("error", err.Error()).Fatal("unable to write secrets file")
		return err
	}

	return nil
}

func (a *environmentAgent) Close() {
	a.stop <- true
}
