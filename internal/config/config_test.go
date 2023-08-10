package appconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	expectedConfig := &AppConfig{
		AuthConfig{
			Host:          "https://www.reddit.com/api/v1/access_token",
			RequestPeriod: 180,
			ClientId:      "testClientId",
			ClientSecret:  "testClientSecret",
			Username:      "Jhon",
			Password:      "Doe",
		},
		ClientConfig{
			Host:          "https://oauth.reddit.com",
			UserAgent:     "dmmakRedditApi/1.0",
			NewPostsUrl:   "/new",
			SavePostUrl:   "/api/save",
			RequestPeriod: 180,
			Subreddits: []Subreddit{
				{
					"golang",
					"slice, map, update, news",
				},
				{
					"wallstreetbets",
					"stock, share, bond",
				},
			},
		},
	}

	os.Setenv("REDDITAPI_CLIENT_ID", "testClientId")
	os.Setenv("REDDITAPI_CLIENT_SECRET", "testClientSecret")
	os.Setenv("REDDITAPI_USERNAME", "Jhon")
	os.Setenv("REDDITAPI_PASSWORD", "Doe")
	LoadConfig("../../testdata/appconfig/test_config.yml")
	actualConfig := cfg

	assert.Equal(t, expectedConfig, actualConfig)

}
