package redditclient

import (
	"context"
	"dmmak/redditapi/internal/api"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type (
	rateLimitedClient struct {
		host            string
		newPostsUrl     string
		savePostUrl     string
		userAgent       string
		rl              *rateLimiter
		authTokenPoller api.AuthTokenPoller
	}
)

const (
	remainingHeader = "x-ratelimit-remaining"
	usedHeader      = "x-ratelimit-used"
	resetHeader     = "x-ratelimit-reset"
)

func NewClient(host string, newPostsUrl string, savePostUrl string, userAgent string,
	authTokenPoller api.AuthTokenPoller) (cl api.RedditAPIClient) {
	cl = &rateLimitedClient{
		host:            host,
		newPostsUrl:     newPostsUrl,
		savePostUrl:     savePostUrl,
		userAgent:       userAgent,
		rl:              &rateLimiter{},
		authTokenPoller: authTokenPoller,
	}
	return cl
}

func (cl *rateLimitedClient) updateRateLimit(resp *http.Response) (err error) {
	remaining, err := strconv.ParseFloat(resp.Header.Get(remainingHeader), 32)
	if err != nil {
		err = fmt.Errorf("error while parsing \"remaining\" ratelimit header: %w", err)
		return err
	}
	used, err := strconv.Atoi(resp.Header.Get(usedHeader))
	if err != nil {
		err = fmt.Errorf("error while parsing \"used\" ratelimit header: %w", err)
		return err
	}
	reset, err := strconv.Atoi(resp.Header.Get(resetHeader))
	if err != nil {
		err = fmt.Errorf("error while parsing \"reset\" ratelimit header: %w", err)
		return err
	}
	cl.rl.update(float32(remaining), used, reset)
	return nil
}

func (cl *rateLimitedClient) setRequestParams(req *http.Request, paramMap map[string]string) {
	authToken := cl.authTokenPoller.TokenValue()
	bearer := "Bearer " + authToken
	req.Header.Add("Authorization", bearer)
	req.Header.Add("User-Agent", cl.userAgent)
	params := req.URL.Query()
	for k, v := range paramMap {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()
}

func (cl *rateLimitedClient) sendApiRequest(ctx context.Context, method string, url string, params map[string]string) (resp *http.Response, err error) {

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		err = fmt.Errorf("error while creating API request: url=%v, params=%v: %w", url, params, err)
		return nil, err
	}
	cl.setRequestParams(req, params)
	// check rate limit params before sending request
	if v := cl.rl.timeToWait(); v > 0 {
		log.Printf("Wait for rate limit resetting for %v seconds\n", v)
		select {
		case <-time.After(time.Duration(v) * time.Second): // block until rate limiter will be resetted at the server side
			log.Printf("Enable API requests")
		case <-ctx.Done():
			err = fmt.Errorf("waiting for rate limit resetting were interrupted")
			return nil, err
		}
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("error while requesting API method: url=%v, params=%v: %w", url, params, err)
		return nil, err
	}

	err = cl.updateRateLimit(resp)
	if err != nil {
		log.Printf("Error updating rate limit value %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid response status code for API request: url=%v, params=%v: statusCode=%v", url, params, resp.StatusCode)
		resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

func (cl *rateLimitedClient) GetNewPosts(ctx context.Context, subreddit, lastPostName string) (newPosts *api.NewPostsResponse, err error) {
	newPosts = &api.NewPostsResponse{}

	url := cl.host + "/r/" + subreddit + cl.newPostsUrl
	params := make(map[string]string)
	params["limit"] = "10"
	params["before"] = lastPostName

	resp, err := cl.sendApiRequest(ctx, http.MethodGet, url, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(newPosts)
	if err != nil {
		err = fmt.Errorf("error while umarshalling new subreddit posts response: %w", err)
		return nil, err
	}
	return newPosts, nil
}

func (cl *rateLimitedClient) SavePost(ctx context.Context, name string) (err error) {
	url := cl.host + cl.savePostUrl
	params := make(map[string]string)
	params["id"] = name

	resp, err := cl.sendApiRequest(ctx, http.MethodPost, url, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
