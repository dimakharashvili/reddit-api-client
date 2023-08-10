package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeAuthRequestBody(t *testing.T) {
	tp := &authTokenPoller{
		url:           "blank",
		requestPeriod: 100,
		creds: Credentials{
			UserName: "Jhon",
			Password: "Doe",
		},
	}
	expected := "grant_type=password&username=Jhon&password=Doe"
	actual, err := ioutil.ReadAll(tp.makeAuthRequestBody())
	if err != nil {
		t.Fatalf("Failed to read request body, err: %v", err)
	}
	assert.Equal(t, expected, string(actual))
}

func TestMakeAuthRequest(t *testing.T) {
	tp := &authTokenPoller{
		url:           "http://reddit.com",
		requestPeriod: 100,
		creds: Credentials{
			UserName:     "Jhon",
			Password:     "Doe",
			ClientId:     "clientId",
			ClientSecret: "clientSecret",
		},
	}
	expectedBody := tp.makeAuthRequestBody()
	expected, err := http.NewRequest(http.MethodPost, "http://reddit.com", expectedBody)
	if err != nil {
		t.Fatalf("Failed to make expected request, err: %v", err)
	}
	expected.SetBasicAuth("clientId", "clientSecret")
	actual, err := tp.makeAuthRequest()
	if err != nil {
		t.Fatalf("Failed to make actual request, err: %v", err)
	}
	assert.Equal(t, expected.Header["Authorization"], actual.Header["Authorization"])
}

func TestRequestAuthToken(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   []byte
	}{
		{
			"success",
			http.StatusOK,
			[]byte(`{"access_token":"someValue","expires_in":180}`),
		},
		{
			"failure",
			http.StatusUnauthorized,
			[]byte(""),
		},
	}
	for _, c := range cases {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			w.Write(c.body)
		}
		ts := httptest.NewServer(http.HandlerFunc(handler))

		poller := &authTokenPoller{url: ts.URL}
		expectedResponse := &authResponse{}
		json.NewDecoder(bytes.NewBuffer(c.body)).Decode(expectedResponse)
		actualResponse, actualErr := poller.requestAuthToken()

		if c.name == "success" {
			if actualErr != nil {
				t.Errorf("Error on success server response: %v", actualErr)
				continue
			}
			assert.Equal(t, expectedResponse, actualResponse)
		}

		if c.name == "failure" {
			expectedErr := fmt.Errorf("invalid auth token response code: %v", http.StatusUnauthorized)
			assert.Equal(t, expectedErr, actualErr)
		}

		ts.Close()
	}
}

func TestAuthPollerStart(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{
			"success",
			http.StatusOK,
		},
		{
			"failure",
			http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			w.Write([]byte("{}"))
		}
		ts := httptest.NewServer(http.HandlerFunc(handler))

		poller := &authTokenPoller{url: ts.URL, requestPeriod: 180}

		ctx, cancel := context.WithCancel(context.Background())
		actualChan, actualErr := poller.Start(ctx)

		if c.name == "success" {
			assert.Nil(t, actualErr)
			assert.NotNil(t, actualChan)
		}

		if c.name == "failure" {
			assert.NotNil(t, actualErr)
		}

		cancel()
		ts.Close()
	}
}
