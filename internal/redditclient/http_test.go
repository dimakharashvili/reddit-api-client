package redditclient

import (
	"bytes"
	"context"
	"dmmak/redditapi/internal/api"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TokenPollerMock struct {
}

func (p *TokenPollerMock) Start(ctx context.Context) (chan struct{}, error) {
	return nil, nil
}

func (p *TokenPollerMock) TokenValue() string {
	return "someValue"
}

func TestSendApiRequest(t *testing.T) {
	cases := []struct {
		name         string
		responseCode int
	}{
		{
			"success",
			http.StatusOK,
		},
		{
			"failureResponseCode",
			http.StatusNotFound,
		},
	}

	for _, c := range cases {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(remainingHeader, "600")
			w.Header().Add(usedHeader, "0")
			w.Header().Add(resetHeader, "600")
			w.WriteHeader(c.responseCode)
			w.Write([]byte("{}"))
		}
		ts := httptest.NewServer(http.HandlerFunc(handler))

		tp := &TokenPollerMock{}
		cl := &rateLimitedClient{
			host:            ts.URL,
			newPostsUrl:     "dummy",
			savePostUrl:     "dummy",
			userAgent:       "dummy",
			rl:              &rateLimiter{},
			authTokenPoller: tp,
		}

		actualResponse, actualErr := cl.sendApiRequest(context.Background(), http.MethodGet, ts.URL, make(map[string]string))
		if c.name == "success" {
			defer actualResponse.Body.Close()
			assert.Nil(t, actualErr)
		}
		if c.name == "failureResponseCode" {
			assert.NotNil(t, actualErr)
		}
		ts.Close()
	}

}

func TestGetNewPosts(t *testing.T) {
	cases := []struct {
		name             string
		responseCode     int
		responseFilePath string
	}{
		{
			"success",
			http.StatusOK,
			"../../testdata/redditclient/successNewPostResponse.json",
		},
		{
			"failureResponseCode",
			http.StatusNotFound,
			"",
		},
	}

	for _, c := range cases {
		responseBytes := readFile(c.responseFilePath, t)

		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(remainingHeader, "600")
			w.Header().Add(usedHeader, "0")
			w.Header().Add(resetHeader, "600")
			w.WriteHeader(c.responseCode)
			w.Write(responseBytes)
		}
		ts := httptest.NewServer(http.HandlerFunc(handler))

		tp := &TokenPollerMock{}
		cl := NewClient(ts.URL, "/dummy", "/dummy", "dummy", tp)

		expectedResponse := &api.NewPostsResponse{}
		json.NewDecoder(bytes.NewBuffer(responseBytes)).Decode(expectedResponse)

		actualResponse, actualErr := cl.GetNewPosts(context.Background(), "golang", "t3_151rq7s")

		if c.name == "success" {
			assert.Nil(t, actualErr)
			assert.Equal(t, expectedResponse, actualResponse)
		}
		if c.name == "failureResponseCode" {
			assert.NotNil(t, actualErr)
		}
		ts.Close()
	}

}

func TestSavePost(t *testing.T) {
	cases := []struct {
		name         string
		responseCode int
	}{
		{
			"success",
			http.StatusOK,
		},
		{
			"failureResponseCode",
			http.StatusNotFound,
		},
	}

	for _, c := range cases {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(remainingHeader, "600")
			w.Header().Add(usedHeader, "0")
			w.Header().Add(resetHeader, "600")
			w.WriteHeader(c.responseCode)
		}
		ts := httptest.NewServer(http.HandlerFunc(handler))

		tp := &TokenPollerMock{}
		cl := NewClient(ts.URL, "/dummy", "/dummy", "dummy", tp)

		actualErr := cl.SavePost(context.Background(), "postName")

		if c.name == "success" {
			assert.Nil(t, actualErr)
		}
		if c.name == "failureResponseCode" {
			assert.NotNil(t, actualErr)
		}
		ts.Close()
	}

}

func readFile(path string, t *testing.T) (b []byte) {
	if path == "" {
		return []byte("{}")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Can't read testing file %v: %v", path, err)
	}
	return b
}
