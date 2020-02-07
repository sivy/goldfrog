package blog

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi"
)

// CreateIndexFunc
func CreateIndexFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating index handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving index...")

		postOpts := GetPostOpts{}
		getPaginationOpts(r, &postOpts)

		isOwner := checkIsOwner(config, r)

		posts := GetPosts(db, postOpts)

		user := User{
			DisplayName: config.Blog.Author.Name,
			Email:       config.Blog.Author.Email,
			Url:         config.Blog.Url,
			Image:       config.Blog.Author.Image,
			IsAdmin:     isOwner,
		}

		post := NewPost(PostOpts{})

		for _, p := range posts {
			p.User = user
		}

		logger.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(config.TemplatesDir, "index.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		logger.Debugf("index template: %v", t)

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			Posts      []*Post
			Post       Post
			Config     Config
			IsOwner    bool
			FormAction string
			TextHeight int
			ShowSlug   bool
			ShowExpand bool
			Flash      string
		}{
			Posts:      posts,
			Post:       post,
			Config:     config,
			IsOwner:    isOwner,
			FormAction: "/new",
			ShowSlug:   false,
			TextHeight: 10,
			ShowExpand: true,
			Flash:      flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

// CreateRssFunc
func CreateRssFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating rss handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving index...")

		postOpts := GetPostOpts{Limit: 10}
		posts := GetPosts(db, postOpts)

		logger.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(config.TemplatesDir, "base/rss.xml")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8" standalone="yes" ?>`))

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "rss", struct {
			Posts  []*Post
			Config Config
			Flash  string
		}{
			Posts:  posts,
			Config: config,
			Flash:  flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

func CreatePostPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating post detail handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving post page...")

		isOwner := checkIsOwner(config, r)

		postSlug := chi.URLParam(r, "slug")

		post, err := GetPostBySlug(db, postSlug)

		user := User{
			DisplayName: config.Blog.Author.Name,
			Email:       config.Blog.Author.Email,
			Url:         config.Blog.Url,
			Image:       config.Blog.Author.Image,
			IsAdmin:     isOwner,
		}

		post.User = user

		logger.Debugf("Found post: %s", post.Title)

		if err != nil {
			logger.Errorf("Could not get post: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		t, err := getTemplate(config.TemplatesDir, "post_detail.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		logger.Debug(t)

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			Post    *Post
			Config  Config
			IsOwner bool
			Flash   string
		}{
			Post:    post,
			Config:  config,
			IsOwner: isOwner,
			Flash:   flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

// CreateIndexFunc
func CreateDailyPostsFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating index handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving index...")
		isOwner := checkIsOwner(config, r)

		year := chi.URLParam(r, "year")
		month := chi.URLParam(r, "month")
		dayOrSlug := chi.URLParam(r, "dayOrSlug")

		isDay, err := regexp.MatchString(`\d+`, dayOrSlug)
		if err != nil {
			logger.Errorf("How did this happen: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		if !isDay {
			logger.Infof("Redirecting old permalink for %s", dayOrSlug)
			post, err := GetPostBySlug(db, dayOrSlug)

			if err != nil {
				logger.Errorf("Could not get post: %v", err)
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, post.PermaLink(), http.StatusPermanentRedirect)
			return
		}

		date, err := time.Parse("2006/01/02",
			fmt.Sprintf("%s/%s/%s", year, month, dayOrSlug))

		if err != nil {
			logger.Errorf("Bad date values: %s, %s, %s", year, month, dayOrSlug)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		posts := GetArchiveDayPosts(db, year, month, dayOrSlug)

		user := User{
			DisplayName: config.Blog.Author.Name,
			Email:       config.Blog.Author.Email,
			Url:         config.Blog.Url,
			Image:       config.Blog.Author.Image,
			IsAdmin:     isOwner,
		}

		for _, p := range posts {
			p.User = user
		}
		logger.Debugf("Found %d posts", len(posts))

		t, err := getTemplate(config.TemplatesDir, "dailydigest.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		logger.Debugf("index template: %v", t)

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			Posts   []*Post
			Post    Post
			Config  Config
			IsOwner bool
			Date    time.Time
			Flash   string
		}{
			Posts:   posts,
			Post:    NewPost(PostOpts{}),
			Config:  config,
			IsOwner: isOwner,
			Date:    date,
			Flash:   flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateArchiveYearMonthFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating main archive list handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving archive year/month...")

		archiveData := GetArchiveYearMonths(db)

		t, err := getTemplate(config.TemplatesDir, "archive_years.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			ArchiveData []ArchiveEntry
			Config      Config
			Flash       string
		}{
			ArchiveData: archiveData,
			Config:      config,
			Flash:       flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateArchivePageFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating archive page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving archive year/month...")
		isOwner := checkIsOwner(config, r)

		// postOpts := GetPostOpts{Limit: 10}
		year := chi.URLParam(r, "year")
		month := chi.URLParam(r, "month")

		posts := GetArchiveMonthPosts(db, year, month)
		user := User{
			DisplayName: config.Blog.Author.Name,
			Email:       config.Blog.Author.Email,
			Url:         config.Blog.Url,
			Image:       config.Blog.Author.Image,
			IsAdmin:     isOwner,
		}

		for _, p := range posts {
			p.User = user
		}
		t, err := getTemplate(config.TemplatesDir, "archive_posts.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			Posts  []*Post
			Config Config
			Year   string
			Month  string
			Flash  string
		}{
			Posts:  posts,
			Config: config,
			Year:   year,
			Month:  month,
			Flash:  flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateSearchPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating search page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving search results...")

		isOwner := checkIsOwner(config, r)

		var posts []*Post
		t, err := getTemplate(config.TemplatesDir, "post_list.html")

		// postOpts := GetPostOpts{Limit: 10}
		term := r.PostFormValue("s")
		if term == "" {
			term = r.URL.Query().Get("s")
		}
		flash, _ := GetFlash(w, r, "flash")

		if term == "" {
			err = t.ExecuteTemplate(w, "base", struct {
				Posts   []*Post
				Config  Config
				Title   string
				IsOwner bool
				Flash   string
			}{
				Posts:   posts,
				Config:  config,
				Title:   fmt.Sprintf("Search"),
				IsOwner: isOwner,
				Flash:   flash,
			})
			return
		}
		opts := GetPostOpts{
			Title: term,
			Body:  term,
		}

		posts = GetPosts(db, opts)
		user := User{
			DisplayName: config.Blog.Author.Name,
			Email:       config.Blog.Author.Email,
			Url:         config.Blog.Url,
			Image:       config.Blog.Author.Image,
			IsAdmin:     isOwner,
		}

		for _, p := range posts {
			p.User = user
		}
		logger.Debugf("found posts: %d", len(posts))

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		err = t.ExecuteTemplate(w, "base", struct {
			Posts   []*Post
			Config  Config
			Title   string
			IsOwner bool
			Flash   string
		}{
			Posts:   posts,
			Config:  config,
			Title:   fmt.Sprintf("Posts found for '%s'", term),
			IsOwner: isOwner,
			Flash:   flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}

func CreateTagPageFunc(config Config, db *sql.DB) http.HandlerFunc {
	logger.Debug("Creating tag page handler")
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving tag search...")

		// postOpts := GetPostOpts{Limit: 10}
		tag := chi.URLParam(r, "tag")

		posts := GetTaggedPosts(db, tag)
		logger.Debugf("tagged posts: %d", len(posts))
		t, err := getTemplate(config.TemplatesDir, "post_list.html")

		if err != nil {
			logger.Errorf("Could not parse template: %v", err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		isOwner := checkIsOwner(config, r)

		flash, _ := GetFlash(w, r, "flash")

		err = t.ExecuteTemplate(w, "base", struct {
			Posts   []*Post
			Config  Config
			Title   string
			IsOwner bool
			Flash   string
		}{
			Posts:   posts,
			Config:  config,
			Title:   fmt.Sprintf("Posts tagged with '%s'", tag),
			IsOwner: isOwner,
			Flash:   flash,
		})

		if err != nil {
			logger.Warnf("Error rendering: %v", err)
		}
	}
}
