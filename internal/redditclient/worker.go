package redditclient

import (
	"context"
	"dmmak/redditapi/internal/api"
	"log"
	"strings"
	"time"
)

type worker struct {
	subreddit     string
	lastPostName  string // keep track of searched posts
	keywords      []string
	requestPeriod uint
	cl            api.RedditAPIClient
}

func NewWorker(subreddit string, keywords string, requestPeriod uint, cl api.RedditAPIClient) (w *worker) {
	splitted := strings.Split(keywords, ",")
	w = &worker{
		subreddit:     subreddit,
		keywords:      splitted,
		requestPeriod: requestPeriod,
		cl:            cl,
	}
	return w
}

func (w *worker) DoWork(ctx context.Context) {
	log.Printf("Start worker for subreddit \"%v\"\n", w.subreddit)
	defer log.Printf("Worker for subreddit \"%v\" is shutted\n", w.subreddit)

	w.saveNewPosts(ctx)
	ticker := time.NewTicker(time.Duration(w.requestPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}
			w.saveNewPosts(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *worker) saveNewPosts(ctx context.Context) {
	newPostsResp, err := w.cl.GetNewPosts(ctx, w.subreddit, w.lastPostName)
	if err != nil {
		log.Printf("Error while getting new posts for subreddit \"%v\": %v\n", w.subreddit, err)
		return
	}

	newPosts := newPostsResp.Data.Children
	if len(newPosts) == 0 {
		log.Printf("No new posts in subreddit %v\n", w.subreddit)
		return
	}
	log.Printf("Found %v new posts in subreddit %v\n", len(newPosts), w.subreddit)

	w.lastPostName = newPosts[0].Data.Name
	for _, post := range newPosts {
		for _, word := range w.keywords {
			if strings.Contains(post.Data.Title, word) {

				log.Printf("Save post id=%v from subreddit \"%v\"\n", post.Data.Name, w.subreddit)
				err = w.cl.SavePost(ctx, post.Data.Name)
				if err != nil {
					log.Printf("Save post id=%v from subreddit \"%v\": %v\n", post.Data.Name, w.subreddit, err)
				}
				break
			}
		}
	}
}
