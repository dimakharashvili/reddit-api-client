package main

import (
	"context"
	"dmmak/redditapi/internal/auth"
	config "dmmak/redditapi/internal/config"
	client "dmmak/redditapi/internal/redditclient"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var configPath string
var logPath string

func setFlags() {
	flag.StringVar(&configPath, "config", "config.yml", "Path to config file")
	flag.StringVar(&logPath, "log", "", "Path to log file")
	flag.Parse()
}

func setLogger(path string) (f *os.File) {
	if path != "" {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		w := io.MultiWriter(os.Stdout, f)
		log.SetOutput(w)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	return f
}

func main() {
	log.Println("Application startup")

	setFlags()
	logFile := setLogger(logPath)
	if logFile != nil {
		defer logFile.Close()
	}

	config.LoadConfig(configPath)
	cfg := config.LoadConfig(configPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	wg := &sync.WaitGroup{}

	tp := auth.NewTokenPoller(cfg.Auth.Host, cfg.Auth.RequestPeriod,
		auth.Credentials{
			UserName:     cfg.Auth.Username,
			Password:     cfg.Auth.Password,
			ClientId:     cfg.Auth.ClientId,
			ClientSecret: cfg.Auth.ClientSecret,
		})
	authExit, err := tp.Start(ctx)
	if err != nil {
		log.Fatalf("Error while starting auth token polling, %v", err)
	}

	cl := client.NewClient(cfg.Client.Host, cfg.Client.NewPostsUrl, cfg.Client.SavePostUrl, cfg.Client.UserAgent, tp)
	// start monitoring subreddits
	for _, sub := range cfg.Client.Subreddits {
		wg.Add(1)
		go func(sub config.Subreddit) {
			defer wg.Done()
			client.NewWorker(sub.Name, sub.Keywords, cfg.Client.RequestPeriod, cl).DoWork(ctx)
		}(sub)
	}

	select {
	case <-authExit:
		cancel()
	case <-ctx.Done():
		log.Println("Shutdown signal recieved")
	}
	wg.Wait()
	<-authExit
	log.Println("Gracefully shutdowned")
}
