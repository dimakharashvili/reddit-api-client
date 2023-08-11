# Reddit API client

## Purpose
This app is created mostly because of educational goals and at the moment it has one implemented feature - looking for subreddits new post titles periodically, matching them with keywords list and saving matched posts to your account. Nothing fancy, but can be useful if you coding your own bots, mods utils, etc.

## Authentication

Simple authentication schema is used and described [here](https://github.com/reddit-archive/reddit/wiki/OAuth2-Quick-Start-Example).\
Auth token is requested periodically and shared among actual API requests.

## Rate limiting
Reddit API obliges users to check rate limit headers after each request.
- X-Ratelimit-Used: Approximate number of requests used in this period
- X-Ratelimit-Remaining: Approximate number of requests left to use
- X-Ratelimit-Reset: Approximate number of seconds to end of periodically

So if you've run out of requests, spamming would be blocked until server resets limits.

## Configuration
config.yml example
```console
auth:
  host: https://www.reddit.com/api/v1/access_token
  requestPeriod: 1800
client:
  host: https://oauth.reddit.com
  userAgent: dmmakRedditApi/1.0
  newPostsUrl: /new
  savePostUrl: /api/save
  requestPeriod: 180
  subreddits:
    - name: golang
      keywords: slice, map, update, news
    - name: wallstreetbets
      keywords: stock, share, bond
```

## Testing and Running

```console
docker run --name reddit-client \
-e REDDITAPI_USERNAME='foo' \
-e REDDITAPI_PASSWORD='bar' \
-e REDDITAPI_CLIENT_ID='id' \
-e REDDITAPI_CLIENT_SECRET='secret' \
-v /some/path/config.yml:/config.yml:ro \
-d reddit-client
```
In case of running on host, flags *--config* and *--log* could be helpful, to set up path to config and log files respectively.






