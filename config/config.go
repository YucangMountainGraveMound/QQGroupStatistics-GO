package config

import (
	"sync"
	"os"
	"gopkg.in/urfave/cli.v1"
	"github.com/BurntSushi/toml"
	"path/filepath"
	"github.com/sirupsen/logrus"
	"bytes"
	"bufio"
	"dormon.net/qq/utils"
)

type tomlConfig struct {
	Secret              string
	SpecificGroup       string
	ElasticSearchConfig elasticSearchConfig
	DatabaseConfig      databaseConfig
	RedisConfig         redisConfig
	Account             []account
	TLS                 tls
}

type elasticSearchConfig struct {
	Host             string
	Port             string
	NumberOfShards   int
	NumberOfReplicas int
	AliasName        string
}

type databaseConfig struct {
	Host         string
	Port         string
	DatabaseName string
	Username     string
	Password     string
}

type redisConfig struct {
	Host string
	Port string
}

type account struct {
	Account string
	Alias   []string
}

type expression struct {
	Name      string
	Character string
}

type tls struct {
	Pem string
	Key string
}

var (
	cfg  *tomlConfig
	once sync.Once
)

func InitialConfig(c *cli.Context) {
	var defaultConfig = tomlConfig{
		Secret:        utils.RandomString(32),
		SpecificGroup: "123456789",
		ElasticSearchConfig: elasticSearchConfig{
			"localhost",
			"9200",
			5,
			1,
			"qq",
		},
		DatabaseConfig: databaseConfig{
			"localhost",
			"5432",
			"name",
			"username",
			"password",
		},
		RedisConfig: redisConfig{
			"localhost",
			"6379",
		},
		Account: []account{
			account{
				"a qq number",
				[]string{"a qq alias"},
			},
			account{
				"another qq number",
				[]string{"a qq alias"},
			},
		},
		TLS: tls{
			Key: "./tls/key.key",
			Pem: "./tls/pem.pem",
		},
	}

	once.Do(func() {
		if !cfgExists(c.String("config")) {
			if c.Bool("generate") {
				configFile, err := os.OpenFile(c.String("config")+".example", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					panic(err)
				}
				var buf bytes.Buffer
				encoder := toml.NewEncoder(&buf)
				err = encoder.Encode(defaultConfig)
				if err != nil {
					panic(err)
				}
				w := bufio.NewWriterSize(configFile, buf.Len())
				if _, err := w.Write(buf.Bytes()); err != nil {
					panic(err)
				}
				if err := w.Flush(); err != nil {
					panic(err)
				}
				defer configFile.Close()
			} else {
				logrus.Fatalf("Cannot locate config file! You can run command [generate] to generate a config file")
				return
			}
		}
		filePath, err := filepath.Abs(c.String("config"))
		if err != nil {
			panic(err)
		}
		if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
			panic(err)
		}
	})
}

func Config() *tomlConfig {
	return cfg
}

func cfgExists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
}
