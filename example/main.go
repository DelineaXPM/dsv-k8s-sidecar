package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

const configDir = "/var/secret/"

func main() {
	logrus.Info("Starting...")
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			b, err := ioutil.ReadFile(configDir + "thy.json")
			if err != nil {
				logrus.WithField("error", err.Error()).Error("Unable to open file")
				return
			}
			var results map[string]interface{}
			if err := json.Unmarshal(b, &results); err != nil {
				logrus.WithField("error", err.Error()).Error("Unable to unmarshal data")
				return
			}

			for key, val := range results {
				logrus.WithFields(logrus.Fields{
					"key":   key,
					"value": val,
				}).Info("Reading secrets")
			}
		}
	}
}
