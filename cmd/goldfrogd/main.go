package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/sivy/goldfrog/pkg/blog"
	"github.com/spf13/viper"
)

var version string // set in linker with ldflags -X main.version=

var log = logrus.New()

func main() {
	log.SetLevel(logrus.DebugLevel)

	var configDir string
	var postsDir string
	var templatesDir string
	var staticDir string
	var uploadsDir string
	var dbFile string
	var showVersionLong bool
	var showVersion bool

	userHomeDir, _ := os.UserHomeDir()
	goldfrogHome, found := os.LookupEnv("BLOGHOME")
	if !found {
		goldfrogHome = filepath.Join(userHomeDir, "goldfrog")
	}

	flag.StringVar(
		&configDir, "config_dir",
		goldfrogHome,
		"Location of config file")

	flag.StringVar(
		&postsDir, "posts_dir",
		goldfrogHome+"/posts",
		"Location of your posts (Jekyll-compatible markdown)")

	flag.StringVar(
		&templatesDir, "templates_dir",
		goldfrogHome+"/templates",
		"Location of blog templates")

	flag.StringVar(
		&staticDir, "static_dir",
		goldfrogHome+"/static",
		"Location of static resources to be served at /static")

	flag.StringVar(
		&uploadsDir, "uploads_dir",
		goldfrogHome+"/uploads",
		"Location of directory to store uploaded files to be served at /uploads")

	flag.StringVar(
		&dbFile, "db",
		goldfrogHome+"/blog.db",
		"File path to sqlite db for indexed content")

	flag.BoolVar(&showVersionLong, "version-long", false, "")
	flag.BoolVar(&showVersion, "version", false, "")
	flag.Parse()

	log.Printf("Using dbFile: %s", dbFile)

	if showVersionLong {
		fmt.Println(version)
		return
	}

	if showVersion {
		tag := strings.Split(version, "-")[0]
		fmt.Println(tag)
		return
	}

	log.Debug("loading config")

	config := loadConfig(configDir)

	if config.PostsDir == "" && postsDir != "" {
		config.PostsDir = postsDir
	}
	if config.TemplatesDir == "" && templatesDir != "" {
		config.TemplatesDir = templatesDir
	}
	if config.StaticDir == "" && staticDir != "" {
		config.StaticDir = staticDir
	}
	if config.UploadsDir == "" && uploadsDir != "" {
		config.UploadsDir = uploadsDir
	}

	log.Debug(postsDir)
	fmt.Println(config.PostsDir)

	if exists, _ := exists(config.PostsDir); exists == false {
		log.Fatalf("PostsDir dir %s does not exist!", config.PostsDir)
	}

	// runWatcher(postsDir, dbFile)
	runServer(config, dbFile)
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func loadConfig(configPath string) blog.Config {
	viper.AddConfigPath(configPath)
	viper.ReadInConfig()

	var config blog.Config
	viper.Unmarshal(&config)

	log.Debugf("config: %v", config)
	return config
}

func runServer(
	config blog.Config, dbFile string) {
	// TODO: config or args with db location and posts dir

	db, err := blog.GetDb(dbFile)
	if err != nil {
		log.Fatalf("Could not get db connection: %v", err)
	}

	repo := blog.PostsRepo{
		PostsDirectory: config.PostsDir,
	}

	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.StripSlashes,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Route("/", func(r chi.Router) {
		r.Mount("/", blog.CreateIndexFunc(config, db))
		r.Mount("/{year}/{month}/{slug}", blog.CreatePostPageFunc(
			config, db))
		r.Mount("/archive", blog.CreateArchiveYearMonthFunc(config, db))
		r.Mount("/archive/{year}/{month}", blog.CreateArchivePageFunc(config, db))
		r.Mount("/tag/{tag}", blog.CreateTagPageFunc(config, db))
		r.Mount("/feed.xml", blog.CreateRssFunc(config, db))
		r.Mount("/search", blog.CreateSearchPageFunc(config, db))

		r.Mount("/new", blog.CreateNewPostFunc(config, db, repo))
		r.Mount(
			"/edit/{postID}",
			blog.CreateEditPostFunc(config, db, repo))
		r.Mount(
			"/edit",
			blog.CreateEditPostFunc(config, db, repo))
		r.Mount("/delete", blog.CreateDeletePostFunc(config, db, repo))

		r.Mount("/signin", blog.CreateSigninPageFunc(config, dbFile))
		r.Mount("/signout", blog.CreateSignoutPageFunc(config, dbFile))

		blog.FileServer(r, "/static", http.Dir(config.StaticDir))
		blog.FileServer(r, "/uploads", http.Dir(config.UploadsDir))
	})

	loc := fmt.Sprintf(fmt.Sprintf(
		"%s:%s", config.Server.Location,
		config.Server.Port))

	log.Info("=====================================")
	log.Infof("Starting GoldFrog on %s", loc)
	log.Info("=====================================")

	http.ListenAndServe(loc, r)
}
