package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/sivy/goldfrog/pkg/blog"
)

var version string // set in linker with ldflags -X main.version=

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

	var postsDir string
	var dbFile string
	var verbose bool
	var showVersionLong bool
	var showVersion bool

	userHomeDir, _ := os.UserHomeDir()
	goldfrogHome, found := os.LookupEnv("BLOGHOME")
	if !found {
		goldfrogHome = filepath.Join(userHomeDir, "goldfrog")
	}

	flag.StringVar(&postsDir, "posts_dir", goldfrogHome+"/posts", "")
	flag.StringVar(&dbFile, "db", goldfrogHome+"/blog.db", "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.BoolVar(&showVersionLong, "version-long", false, "")
	flag.BoolVar(&showVersion, "version", false, "")

	flag.Parse()
	if showVersionLong {
		fmt.Println(version)
		return
	}

	if showVersion {
		tag := strings.Split(version, "-")[0]
		fmt.Println(tag)
		return
	}

	fmt.Printf("Indexing %s", postsDir)

	blog.IndexPosts(postsDir, dbFile, verbose)
}
