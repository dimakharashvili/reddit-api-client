package api

import "context"

type (
	AuthTokenPoller interface {
		Start(ctx context.Context) (exit chan struct{}, err error)
		TokenValue() string
	}

	RedditAPIClient interface {
		GetNewPosts(ctx context.Context, subreddit, lastPostName string) (r *NewPostsResponse, err error)
		SavePost(ctx context.Context, name string) error
	}

	NewPostsResponse struct {
		Data NewPostsResponseData `json:"data"`
	}

	NewPostsResponseData struct {
		Children []NewPostsResponseChildren `json:"children"`
	}

	NewPostsResponseChildren struct {
		Data NewPostsResponseChildrenData `json:"data"`
	}

	NewPostsResponseChildrenData struct {
		Title string `json:"title"`
		Name  string `json:"name"`
	}
)
