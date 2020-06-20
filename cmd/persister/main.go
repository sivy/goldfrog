package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sirupsen/logrus"
	"github.com/sivy/goldfrog/pkg/blog"
)

var version string // set in linker with ldflags -X main.version=

var logger = logrus.New()

func main() {
	logger.SetLevel(logrus.DebugLevel)

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

	flag.StringVar(
		&postsDir, "posts_dir",
		goldfrogHome+"/posts",
		"Location of your posts (Jekyll-compatible markdown)")

	flag.StringVar(
		&dbFile, "db",
		goldfrogHome+"/blog.db",
		"File path to sqlite db for indexed content")

	flag.BoolVar(
		&verbose, "v", false,
		"Enable verbose logging")

	flag.BoolVar(
		&showVersion, "version", false,
		"Print the version")

	flag.BoolVar(
		&showVersionLong, "version-long", false,
		"Print the version (version + git sha1)")

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

	logger.Infof("Persisting posts in db: %s to: %s", dbFile, postsDir)
	repo := blog.FilePostsRepo{
		PostsDirectory: postsDir,
	}

	db, err := blog.GetDb(dbFile)
	if err != nil {
		log.Error(err)
		return
	}
	dbs := blog.NewSqliteStorage(db)

	posts := dbs.GetPosts(blog.GetPostOpts{
		Limit: -1,
	})

	for _, post := range posts {
		err := repo.SavePostFile(post)
		if err != nil {
			log.Error(err)
			return
		}
	}
}
