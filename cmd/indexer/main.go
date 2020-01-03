package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/sivy/goldfrog/pkg/blog"
)

var log = logrus.New()

func runWatcher(postsDir string, dbFile string) {
	// TODO: https://godoc.org/github.com/fsnotify/fsnotify#example-NewWatcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(postsDir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func main() {
	log.SetLevel(logrus.DebugLevel)
	log.Debug("Goldfrog indexer")

	var postsDir string
	var dbFile string
	userHomeDir, _ := os.UserHomeDir()
	goldfrogHome, found := os.LookupEnv("BLOGHOME")
	if !found {
		goldfrogHome = filepath.Join(userHomeDir, "goldfrog")
	}

	flag.StringVar(&postsDir, "posts_dir", goldfrogHome+"/posts", "")
	flag.StringVar(&dbFile, "db", goldfrogHome+"/blog.db", "")
	flag.Parse()
	log.Debug(postsDir)
	fmt.Println(postsDir)

	blog.IndexPosts(postsDir, dbFile)
}
