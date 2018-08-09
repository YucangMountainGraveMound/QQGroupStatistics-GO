package config

import (
	"io/ioutil"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"sync"
)

var exp map[string]string
var expOnce sync.Once

func GetExpression() map[string]string {
	expOnce.Do(func() {
		file, err := ioutil.ReadFile("./expression.json")
		if err != nil {
			logrus.Fatalf("Failed to load expression.json file with error: %s", err)
		}
		json.Unmarshal(file, &exp)
	})

	return exp
}
