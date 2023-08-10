package appconfig

import (
	"fmt"
	"log"
	"os"
	"sync"

	yamlV3 "gopkg.in/yaml.v3"
)

type (
	AppConfig struct {
		Auth   AuthConfig   `yaml:"auth"`
		Client ClientConfig `yaml:"client"`
	}

	AuthConfig struct {
		Host          string `yaml:"host"`
		RequestPeriod uint   `yaml:"requestPeriod"`
		ClientId      string
		ClientSecret  string
		Username      string
		Password      string
	}

	ClientConfig struct {
		Host          string      `yaml:"host"`
		UserAgent     string      `yaml:"userAgent"`
		NewPostsUrl   string      `yaml:"newPostsUrl"`
		SavePostUrl   string      `yaml:"savePostUrl"`
		RequestPeriod uint        `yaml:"requestPeriod"`
		Subreddits    []Subreddit `yaml:"subreddits"`
	}

	Subreddit struct {
		Name     string `yaml:"name"`
		Keywords string `yaml:"keywords"`
	}
)

var cfg *AppConfig
var once sync.Once

func LoadConfig(path string) *AppConfig {
	fl := func() {
		load(path)
	}
	once.Do(fl)
	return cfg
}

func load(path string) {
	f, err := os.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("couldn't read config file: %w", err)
		log.Fatalln(err)
	}
	cfg = &AppConfig{}
	err = yamlV3.Unmarshal(f, cfg)
	if err != nil {
		err = fmt.Errorf("couldn't unmarshall config file: %w", err)
		log.Fatalln(err)
	}
	cfg.Auth.ClientId = os.Getenv("REDDITAPI_CLIENT_ID")
	cfg.Auth.ClientSecret = os.Getenv("REDDITAPI_CLIENT_SECRET")
	cfg.Auth.Username = os.Getenv("REDDITAPI_USERNAME")
	cfg.Auth.Password = os.Getenv("REDDITAPI_PASSWORD")

	log.Println("Config loaded")
}
