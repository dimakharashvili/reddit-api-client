package redditclient

import (
	"context"
	. "dmmak/redditapi/internal/api"
	"testing"
)

type StubRedditAPIClient struct {
}

func (cl *StubRedditAPIClient) GetNewPosts(ctx context.Context, subreddit, lastPostName string) (r *NewPostsResponse, err error) {
	r = &NewPostsResponse{
		Data: NewPostsResponseData{
			Children: []NewPostsResponseChildren{
				{
					Data: NewPostsResponseChildrenData{
						Title: "postTitle1",
						Name:  "postName1",
					},
				},
				{
					Data: NewPostsResponseChildrenData{
						Title: "postTitle1",
						Name:  "postName1",
					},
				},
			},
		},
	}
	return r, nil
}

func (cl *StubRedditAPIClient) SavePost(ctx context.Context, name string) (err error) {
	return nil
}

func TestDoWork(t *testing.T) {
	cl := &StubRedditAPIClient{}
	ctx, cancel := context.WithCancel(context.Background())
	worker := NewWorker("subreddit1", "postTitle1, postTitle2, dummy", 180, cl)
	cancel()
	worker.DoWork(ctx)
}
