package auth

import (
	"bytes"
	"context"
	"dmmak/redditapi/internal/api"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	authTokenPoller struct {
		token         authToken
		url           string
		requestPeriod uint
		creds         Credentials
		mu            sync.RWMutex
	}

	authToken struct {
		value    string
		lifetime int
	}

	Credentials struct {
		UserName     string
		Password     string
		ClientId     string
		ClientSecret string
	}

	authResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
)

func NewTokenPoller(url string, requestPeriod uint, creds Credentials) (tp api.AuthTokenPoller) {
	tp = &authTokenPoller{url: url, requestPeriod: requestPeriod, creds: creds}
	return tp
}

func (p *authTokenPoller) TokenValue() (value string) {
	p.mu.RLock()
	value = p.token.value
	p.mu.RUnlock()
	return value
}

func (p *authTokenPoller) Start(ctx context.Context) (exit chan struct{}, err error) {
	exit = make(chan struct{})
	err = p.refreshAuthToken()
	if err != nil {
		return nil, err
	}
	go func() {
		p.maintainAuthToken(ctx, exit)
		close(exit)
	}()
	log.Println("Auth token poller started")
	return exit, nil
}

func (p *authTokenPoller) refreshAuthToken() (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	authResp, err := p.requestAuthToken()
	if err != nil {
		return err
	}
	p.token.value = authResp.AccessToken
	p.token.lifetime = authResp.ExpiresIn
	return nil
}

func (p *authTokenPoller) maintainAuthToken(ctx context.Context, exit chan<- struct{}) {
	for {
		select {
		case <-time.After(time.Duration(p.requestPeriod) * time.Second):
			err := p.refreshAuthToken()
			if err != nil {
				log.Printf("Error while refreshing auth token: %v\n", err)
				exit <- struct{}{}
				return
			}
		case <-ctx.Done():
			exit <- struct{}{}
			return
		}
	}
}

func (p *authTokenPoller) makeAuthRequest() (req *http.Request, err error) {
	body := p.makeAuthRequestBody()
	req, err = http.NewRequest(http.MethodPost, p.url, body)
	if err != nil {
		return req, err
	}
	req.SetBasicAuth(p.creds.ClientId, p.creds.ClientSecret)
	return req, nil
}

func (p *authTokenPoller) makeAuthRequestBody() (body *bytes.Buffer) {
	var authParams strings.Builder
	authParams.WriteString("grant_type=password")
	authParams.WriteString("&username=")
	authParams.WriteString(p.creds.UserName)
	authParams.WriteString("&password=")
	authParams.WriteString(p.creds.Password)
	body = bytes.NewBuffer([]byte(authParams.String()))
	return body
}

func (p *authTokenPoller) requestAuthToken() (authResp *authResponse, err error) {
	authResp = &authResponse{}
	req, err := p.makeAuthRequest()
	if err != nil {
		return authResp, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return authResp, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid auth token response code: %d", resp.StatusCode)
		return nil, err
	}
	err = json.NewDecoder(resp.Body).Decode(authResp)
	if err != nil {
		err = fmt.Errorf("error decoding auth response: %w", err)
		return nil, err
	}
	return authResp, nil
}
